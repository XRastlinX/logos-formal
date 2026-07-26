import json
import stat
import sys
import tempfile
import unittest
import zipfile
from argparse import Namespace
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import bridge


class BridgeTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.root = Path(self.temp.name)
        self.task = self.root / "task.md"
        self.source = self.root / "source.md"
        self.task.write_text("# Task\nReview exactly.\n", encoding="utf-8")
        self.source.write_text("# Source\nBounded artifact.\n", encoding="utf-8")

    def tearDown(self):
        self.temp.cleanup()

    def pack(self, name="packet.zip"):
        output = self.root / name
        result = bridge.command_pack(
            Namespace(
                task=str(self.task),
                include=[str(self.source)],
                output=str(output),
            )
        )
        return output, result

    def test_valid_packet_verifies(self):
        packet, result = self.pack()
        verified = bridge.command_verify_export(Namespace(archive=str(packet)))
        self.assertEqual(result["packetId"], verified["packetId"])
        self.assertEqual("NONE", verified["authorityEffect"])

    def test_pack_is_byte_deterministic(self):
        first, first_result = self.pack("first.zip")
        second, second_result = self.pack("second.zip")
        self.assertEqual(first.read_bytes(), second.read_bytes())
        self.assertEqual(first_result["packetId"], second_result["packetId"])

    def test_payload_tamper_is_rejected(self):
        packet, _ = self.pack()
        entries = bridge.read_safe_zip(packet)
        entries["payload/source.md"] = b"tampered"
        tampered = self.root / "tampered.zip"
        bridge.write_deterministic_zip(tampered, entries)
        with self.assertRaisesRegex(bridge.BridgeError, "PAYLOAD_SIZE_MISMATCH|PAYLOAD_HASH_MISMATCH"):
            bridge.command_verify_export(Namespace(archive=str(tampered)))

    def test_unmanifested_file_is_rejected(self):
        packet, _ = self.pack()
        entries = bridge.read_safe_zip(packet)
        entries["extra.md"] = b"unexpected"
        bad = self.root / "extra.zip"
        bridge.write_deterministic_zip(bad, entries)
        with self.assertRaisesRegex(bridge.BridgeError, "UNMANIFESTED_PATH"):
            bridge.command_verify_export(Namespace(archive=str(bad)))

    def test_path_traversal_is_rejected(self):
        bad = self.root / "traversal.zip"
        with zipfile.ZipFile(bad, "w") as archive:
            archive.writestr("../escape.md", "x")
        with self.assertRaisesRegex(bridge.BridgeError, "UNSAFE_PATH"):
            bridge.read_safe_zip(bad)

    def test_drive_path_is_rejected(self):
        bad = self.root / "drive.zip"
        with zipfile.ZipFile(bad, "w") as archive:
            archive.writestr("C:/escape.md", "x")
        with self.assertRaisesRegex(bridge.BridgeError, "DRIVE_PATH_FORBIDDEN"):
            bridge.read_safe_zip(bad)

    def test_symlink_is_rejected(self):
        bad = self.root / "symlink.zip"
        with zipfile.ZipFile(bad, "w") as archive:
            info = zipfile.ZipInfo("link.md")
            info.create_system = 3
            info.external_attr = (stat.S_IFLNK | 0o777) << 16
            archive.writestr(info, "target")
        with self.assertRaisesRegex(bridge.BridgeError, "SYMLINK_FORBIDDEN"):
            bridge.read_safe_zip(bad)

    def test_private_key_content_is_rejected(self):
        self.source.write_text(
            "-----BEGIN PRIVATE KEY-----\nnot-real\n", encoding="utf-8"
        )
        with self.assertRaisesRegex(bridge.BridgeError, "SECRET_SHAPED_CONTENT"):
            self.pack()

    def make_response(self, packet_id, response_id=None):
        folder = self.root / ("response-" + str(len(list(self.root.glob("response-*")))))
        folder.mkdir()
        (folder / "INPUT_PACKET_ID.txt").write_text(
            response_id or packet_id, encoding="utf-8"
        )
        (folder / "response.md").write_text(
            "# Review\nNo authority claimed.\n", encoding="utf-8"
        )
        return folder

    def test_valid_response_is_quarantined(self):
        _, packed = self.pack()
        response = self.make_response(packed["packetId"])
        result = bridge.command_ingest_response(
            Namespace(
                response=str(response),
                expected_packet_id=packed["packetId"],
                stage_dir=str(self.root / "quarantine"),
            )
        )
        self.assertEqual("QUARANTINED_UNTRUSTED", result["disposition"])
        receipt = Path(result["stagePath"]) / "GROK_RESPONSE_INTAKE_RECEIPT.json"
        self.assertTrue(receipt.is_file())
        body = json.loads(receipt.read_text(encoding="utf-8"))
        self.assertEqual("NONE", body["authorityEffect"])

    def test_wrong_response_binding_is_rejected(self):
        _, packed = self.pack()
        response = self.make_response(packed["packetId"], "sha256:" + "0" * 64)
        with self.assertRaisesRegex(bridge.BridgeError, "RESPONSE_PACKET_BINDING_MISMATCH"):
            bridge.command_ingest_response(
                Namespace(
                    response=str(response),
                    expected_packet_id=packed["packetId"],
                    stage_dir=str(self.root / "quarantine"),
                )
            )

    def test_response_outside_contract_is_rejected(self):
        _, packed = self.pack()
        response = self.make_response(packed["packetId"])
        (response / "surprise.md").write_text("x", encoding="utf-8")
        with self.assertRaisesRegex(bridge.BridgeError, "RESPONSE_PATH_OUTSIDE_CONTRACT"):
            bridge.command_ingest_response(
                Namespace(
                    response=str(response),
                    expected_packet_id=packed["packetId"],
                    stage_dir=str(self.root / "quarantine"),
                )
            )

    def test_response_secret_is_rejected(self):
        _, packed = self.pack()
        response = self.make_response(packed["packetId"])
        (response / "response.md").write_text(
            "-----BEGIN OPENSSH PRIVATE KEY-----\nnot-real", encoding="utf-8"
        )
        with self.assertRaisesRegex(bridge.BridgeError, "SECRET_SHAPED_CONTENT"):
            bridge.command_ingest_response(
                Namespace(
                    response=str(response),
                    expected_packet_id=packed["packetId"],
                    stage_dir=str(self.root / "quarantine"),
                )
            )

    def test_response_nested_archive_is_rejected(self):
        _, packed = self.pack()
        response = self.make_response(packed["packetId"])
        artifacts = response / "artifacts"
        artifacts.mkdir()
        (artifacts / "nested.zip").write_bytes(b"PK")
        with self.assertRaisesRegex(bridge.BridgeError, "FORBIDDEN_FILE_TYPE"):
            bridge.command_ingest_response(
                Namespace(
                    response=str(response),
                    expected_packet_id=packed["packetId"],
                    stage_dir=str(self.root / "quarantine"),
                )
            )


if __name__ == "__main__":
    unittest.main()
