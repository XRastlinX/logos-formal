#!/usr/bin/env python3
"""Fail-closed portable collaboration bridge for a remote Grok sandbox.

The bridge has two deliberately asymmetric operations:

* ``pack`` exports an explicitly selected, content-addressed review packet.
* ``ingest-response`` imports a bounded response into local quarantine.

Neither operation grants Owner, Seal, Apply, or state-commit authority.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import shutil
import stat
import sys
import tempfile
import zipfile
from pathlib import Path, PurePosixPath
from typing import Any, Iterable


SCHEMA = "ark.grok-collaboration-packet"
SCHEMA_VERSION = "1.0.0"
RECEIPT_SCHEMA = "ark.grok-response-intake-receipt"
CANONICAL_PROFILE = "ARK_CANONICAL_JSON_SUBSET_V1"
GOVERNANCE_STATE = "010"
AUTHORITY_EFFECT = "NONE"
MODE = "REVIEW_ONLY"
REMOTE_DROP = "/home/workdir/artifacts"

MAX_FILES = 32
MAX_FILE_BYTES = 5 * 1024 * 1024
MAX_TOTAL_BYTES = 20 * 1024 * 1024

ALLOWED_SUFFIXES = {
    ".css",
    ".csv",
    ".html",
    ".js",
    ".json",
    ".jsx",
    ".md",
    ".mjs",
    ".py",
    ".toml",
    ".ts",
    ".tsx",
    ".txt",
    ".yaml",
    ".yml",
}

FORBIDDEN_BASENAMES = {
    ".env",
    ".git-credentials",
    "credentials.json",
    "id_dsa",
    "id_ed25519",
    "id_rsa",
    "known_hosts",
}

HIGH_CONFIDENCE_SECRET_PATTERNS = (
    re.compile(rb"-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----"),
    re.compile(rb"(?i)authorization\s*:\s*bearer\s+[A-Za-z0-9._~+/=-]{16,}"),
    re.compile(rb"\bAKIA[0-9A-Z]{16}\b"),
    re.compile(rb"\bgh[pousr]_[A-Za-z0-9]{30,}\b"),
)


class BridgeError(RuntimeError):
    """Typed fail-closed bridge error."""


def canonical_bytes(value: Any) -> bytes:
    """Encode the manifest subset deterministically.

    The bridge forbids floating-point values in signed/hash-bound structures.
    This is a deliberately small deterministic JSON profile, not a claim that
    Python's encoder implements all RFC 8785 edge cases.
    """

    def reject_floats(node: Any, path: str = "$") -> None:
        if isinstance(node, float):
            raise BridgeError(f"FLOAT_FORBIDDEN:{path}")
        if isinstance(node, dict):
            for key, child in node.items():
                if not isinstance(key, str):
                    raise BridgeError(f"NON_STRING_KEY:{path}")
                reject_floats(child, f"{path}.{key}")
        elif isinstance(node, list):
            for index, child in enumerate(node):
                reject_floats(child, f"{path}[{index}]")

    reject_floats(value)
    return json.dumps(
        value,
        ensure_ascii=False,
        allow_nan=False,
        sort_keys=True,
        separators=(",", ":"),
    ).encode("utf-8")


def sha256_bytes(data: bytes) -> str:
    return "sha256:" + hashlib.sha256(data).hexdigest()


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return "sha256:" + digest.hexdigest()


def safe_archive_name(name: str) -> str:
    normalized = name.replace("\\", "/")
    path = PurePosixPath(normalized)
    if not normalized or normalized.startswith("/") or path.is_absolute():
        raise BridgeError(f"UNSAFE_PATH:{name}")
    if re.match(r"^[A-Za-z]:", normalized):
        raise BridgeError(f"DRIVE_PATH_FORBIDDEN:{name}")
    if any(part in {"", ".", ".."} for part in path.parts):
        raise BridgeError(f"UNSAFE_PATH_SEGMENT:{name}")
    if "\x00" in normalized:
        raise BridgeError(f"NUL_IN_PATH:{name}")
    return path.as_posix()


def validate_payload_name(name: str) -> None:
    archive_name = safe_archive_name(name)
    basename = PurePosixPath(archive_name).name.casefold()
    if basename in FORBIDDEN_BASENAMES:
        raise BridgeError(f"FORBIDDEN_FILENAME:{name}")
    suffix = PurePosixPath(archive_name).suffix.casefold()
    if suffix not in ALLOWED_SUFFIXES:
        raise BridgeError(f"FORBIDDEN_FILE_TYPE:{name}")


def scan_for_secrets(name: str, data: bytes) -> None:
    for pattern in HIGH_CONFIDENCE_SECRET_PATTERNS:
        if pattern.search(data):
            raise BridgeError(f"SECRET_SHAPED_CONTENT:{name}")


def validate_size_set(items: Iterable[tuple[str, bytes]]) -> list[tuple[str, bytes]]:
    materialized = list(items)
    if len(materialized) > MAX_FILES:
        raise BridgeError(f"FILE_COUNT_EXCEEDED:{len(materialized)}")
    total = 0
    casefold_names: set[str] = set()
    for name, data in materialized:
        safe_name = safe_archive_name(name)
        folded = safe_name.casefold()
        if folded in casefold_names:
            raise BridgeError(f"DUPLICATE_CASEFOLD_PATH:{name}")
        casefold_names.add(folded)
        if len(data) > MAX_FILE_BYTES:
            raise BridgeError(f"FILE_SIZE_EXCEEDED:{name}")
        total += len(data)
    if total > MAX_TOTAL_BYTES:
        raise BridgeError(f"TOTAL_SIZE_EXCEEDED:{total}")
    return materialized


def zip_info(name: str) -> zipfile.ZipInfo:
    info = zipfile.ZipInfo(safe_archive_name(name), date_time=(1980, 1, 1, 0, 0, 0))
    info.compress_type = zipfile.ZIP_STORED
    info.create_system = 3
    info.external_attr = (stat.S_IFREG | 0o644) << 16
    info.flag_bits |= 0x800
    return info


def write_deterministic_zip(output: Path, entries: dict[str, bytes]) -> None:
    output.parent.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(output, "w") as archive:
        for name in sorted(entries):
            archive.writestr(zip_info(name), entries[name])


def read_safe_zip(path: Path) -> dict[str, bytes]:
    if not path.is_file():
        raise BridgeError(f"ARCHIVE_NOT_FOUND:{path}")
    entries: dict[str, bytes] = {}
    with zipfile.ZipFile(path, "r") as archive:
        infos = archive.infolist()
        if len(infos) > MAX_FILES:
            raise BridgeError(f"FILE_COUNT_EXCEEDED:{len(infos)}")
        for info in infos:
            name = safe_archive_name(info.filename)
            unix_mode = (info.external_attr >> 16) & 0xFFFF
            if stat.S_ISLNK(unix_mode):
                raise BridgeError(f"SYMLINK_FORBIDDEN:{name}")
            if info.is_dir():
                raise BridgeError(f"DIRECTORY_ENTRY_FORBIDDEN:{name}")
            if info.file_size > MAX_FILE_BYTES:
                raise BridgeError(f"FILE_SIZE_EXCEEDED:{name}")
            folded = name.casefold()
            if any(existing.casefold() == folded for existing in entries):
                raise BridgeError(f"DUPLICATE_CASEFOLD_PATH:{name}")
            entries[name] = archive.read(info)
    return dict(validate_size_set(entries.items()))


def read_response_source(path: Path) -> dict[str, bytes]:
    if path.is_file():
        return read_safe_zip(path)
    if not path.is_dir():
        raise BridgeError(f"RESPONSE_SOURCE_NOT_FOUND:{path}")
    items: list[tuple[str, bytes]] = []
    for candidate in sorted(path.rglob("*")):
        if candidate.is_symlink():
            raise BridgeError(f"SYMLINK_FORBIDDEN:{candidate}")
        if candidate.is_file():
            relative = candidate.relative_to(path).as_posix()
            items.append((safe_archive_name(relative), candidate.read_bytes()))
    return dict(validate_size_set(items))


def packet_core(task_hash: str, return_contract_hash: str, payload: list[dict[str, Any]]) -> dict[str, Any]:
    return {
        "schema": SCHEMA,
        "schemaVersion": SCHEMA_VERSION,
        "canonicalizationProfile": CANONICAL_PROFILE,
        "mode": MODE,
        "governanceState": GOVERNANCE_STATE,
        "authorityEffect": AUTHORITY_EFFECT,
        "sourceStore": "CODEX_LOCAL_EXPLICIT_EXPORT",
        "intendedRecipient": "GROK_REMOTE_SANDBOX",
        "intendedDropLocation": REMOTE_DROP,
        "taskHash": task_hash,
        "returnContractHash": return_contract_hash,
        "payload": payload,
        "boundaries": {
            "ownerAuthorityTransferred": False,
            "applyAuthorityTransferred": False,
            "credentialsIncluded": False,
            "remoteExecutionAuthorized": False,
            "responseAutoApplied": False,
        },
    }


def build_return_contract(packet_id: str) -> dict[str, Any]:
    return {
        "schema": "ark.grok-collaboration-return-contract",
        "schemaVersion": "1.0.0",
        "inputPacketId": packet_id,
        "requiredFiles": ["INPUT_PACKET_ID.txt", "response.md"],
        "optionalPathPrefix": "artifacts/",
        "allowedFileTypes": sorted(ALLOWED_SUFFIXES),
        "maximumFiles": MAX_FILES,
        "maximumFileBytes": MAX_FILE_BYTES,
        "maximumTotalBytes": MAX_TOTAL_BYTES,
        "requiredAuthorityEffect": AUTHORITY_EFFECT,
        "requiredGovernanceState": GOVERNANCE_STATE,
        "ingestDisposition": "QUARANTINED_UNTRUSTED",
        "prohibitions": [
            "Do not include credentials, cookies, access tokens, or private keys.",
            "Do not claim Owner, Seal, Apply, or STATE_COMMIT authority.",
            "Do not request direct access to the sender's local drives.",
            "Do not return executable binaries or archive-within-archive payloads.",
        ],
    }


def build_remote_instructions(packet_id: str) -> str:
    return f"""# Grok collaboration instructions

This is a bounded review packet. It does not grant filesystem, Owner, Seal,
Apply, or state-commit authority.

1. Read `task.md`, `grok-exchange-manifest.json`, and the files under `payload/`.
2. Perform only the review or analysis requested in `task.md`.
3. Place the exact text below in `INPUT_PACKET_ID.txt`:

   `{packet_id}`

4. Put the principal answer in `response.md`.
5. Optional supporting text/data files may be placed under `artifacts/`.
6. Return those files as a folder or ZIP. Do not include credentials, cookies,
   tokens, private keys, executables, or nested archives.

The returned material will be treated as untrusted external input, verified,
content-addressed, and staged in quarantine. It will never be auto-applied.

Required response metadata:

```text
governance_state: 010
authority_effect: NONE
status: EXTERNAL_REVIEW
```
"""


def command_pack(args: argparse.Namespace) -> dict[str, Any]:
    task_path = Path(args.task).resolve()
    if not task_path.is_file():
        raise BridgeError(f"TASK_NOT_FOUND:{task_path}")
    task_data = task_path.read_bytes()
    validate_payload_name(task_path.name)
    scan_for_secrets("task.md", task_data)

    payload_entries: list[dict[str, Any]] = []
    payload_bytes: dict[str, bytes] = {}
    seen_names: set[str] = set()
    for raw_path in args.include:
        source = Path(raw_path).resolve()
        if not source.is_file() or source.is_symlink():
            raise BridgeError(f"INPUT_NOT_REGULAR_FILE:{source}")
        archive_name = safe_archive_name(f"payload/{source.name}")
        validate_payload_name(archive_name)
        folded = archive_name.casefold()
        if folded in seen_names:
            raise BridgeError(f"PAYLOAD_BASENAME_COLLISION:{source.name}")
        seen_names.add(folded)
        data = source.read_bytes()
        if len(data) > MAX_FILE_BYTES:
            raise BridgeError(f"FILE_SIZE_EXCEEDED:{source.name}")
        scan_for_secrets(archive_name, data)
        payload_bytes[archive_name] = data
        payload_entries.append(
            {
                "path": archive_name,
                "sourceLabel": source.name,
                "bytes": len(data),
                "sha256": sha256_bytes(data),
            }
        )

    payload_entries.sort(key=lambda item: item["path"])
    provisional_contract = build_return_contract("PENDING")
    provisional_contract_hash = sha256_bytes(canonical_bytes(provisional_contract))
    core = packet_core(
        sha256_bytes(task_data),
        provisional_contract_hash,
        payload_entries,
    )
    packet_id = sha256_bytes(canonical_bytes(core))

    return_contract = build_return_contract(packet_id)
    # The contract points back to packetId. To avoid a circular digest, the
    # packet core binds the contract's normalized form with inputPacketId set to
    # PENDING; verification performs the same normalization.
    manifest = dict(core)
    manifest["packetId"] = packet_id
    manifest["packetRootRule"] = (
        "sha256(canonical(packetCore)); returnContractHash uses inputPacketId=PENDING"
    )

    entries = {
        "GROK_INSTRUCTIONS.md": build_remote_instructions(packet_id).encode("utf-8"),
        "grok-exchange-manifest.json": canonical_bytes(manifest) + b"\n",
        "return-contract.json": canonical_bytes(return_contract) + b"\n",
        "task.md": task_data,
        **payload_bytes,
    }
    validate_size_set(entries.items())
    output = Path(args.output).resolve()
    write_deterministic_zip(output, entries)

    result = {
        "status": "PACKET_CREATED",
        "packetId": packet_id,
        "archive": str(output),
        "archiveSha256": sha256_file(output),
        "payloadFiles": len(payload_entries),
        "authorityEffect": AUTHORITY_EFFECT,
        "governanceState": GOVERNANCE_STATE,
    }
    return result


def normalized_return_contract_hash(contract: dict[str, Any]) -> str:
    normalized = dict(contract)
    normalized["inputPacketId"] = "PENDING"
    return sha256_bytes(canonical_bytes(normalized))


def command_verify_export(args: argparse.Namespace) -> dict[str, Any]:
    archive_path = Path(args.archive).resolve()
    entries = read_safe_zip(archive_path)
    required = {
        "GROK_INSTRUCTIONS.md",
        "grok-exchange-manifest.json",
        "return-contract.json",
        "task.md",
    }
    missing = required - entries.keys()
    if missing:
        raise BridgeError("MISSING_REQUIRED:" + ",".join(sorted(missing)))
    for name in entries:
        if name not in required and not name.startswith("payload/"):
            raise BridgeError(f"UNMANIFESTED_PATH:{name}")
    manifest = json.loads(entries["grok-exchange-manifest.json"])
    contract = json.loads(entries["return-contract.json"])
    if manifest.get("schema") != SCHEMA or manifest.get("schemaVersion") != SCHEMA_VERSION:
        raise BridgeError("MANIFEST_SCHEMA_MISMATCH")
    if manifest.get("authorityEffect") != AUTHORITY_EFFECT:
        raise BridgeError("AUTHORITY_ELEVATION")
    if manifest.get("governanceState") != GOVERNANCE_STATE:
        raise BridgeError("GOVERNANCE_STATE_MISMATCH")
    if manifest.get("mode") != MODE:
        raise BridgeError("MODE_MISMATCH")
    if manifest.get("taskHash") != sha256_bytes(entries["task.md"]):
        raise BridgeError("TASK_HASH_MISMATCH")
    packet_id = manifest.get("packetId")
    if not isinstance(packet_id, str):
        raise BridgeError("PACKET_ID_MISSING")
    if contract.get("inputPacketId") != packet_id:
        raise BridgeError("RETURN_CONTRACT_PACKET_MISMATCH")
    if normalized_return_contract_hash(contract) != manifest.get("returnContractHash"):
        raise BridgeError("RETURN_CONTRACT_HASH_MISMATCH")
    expected_payload_paths = set()
    for item in manifest.get("payload", []):
        name = safe_archive_name(item["path"])
        expected_payload_paths.add(name)
        if name not in entries:
            raise BridgeError(f"PAYLOAD_MISSING:{name}")
        if item.get("bytes") != len(entries[name]):
            raise BridgeError(f"PAYLOAD_SIZE_MISMATCH:{name}")
        if item.get("sha256") != sha256_bytes(entries[name]):
            raise BridgeError(f"PAYLOAD_HASH_MISMATCH:{name}")
    actual_payload_paths = {name for name in entries if name.startswith("payload/")}
    if actual_payload_paths != expected_payload_paths:
        raise BridgeError("PAYLOAD_SET_MISMATCH")
    core = dict(manifest)
    core.pop("packetId", None)
    core.pop("packetRootRule", None)
    if sha256_bytes(canonical_bytes(core)) != packet_id:
        raise BridgeError("PACKET_ID_MISMATCH")
    for name, data in entries.items():
        if name == "grok-exchange-manifest.json":
            continue
        scan_for_secrets(name, data)
    return {
        "status": "PACKET_VERIFIED",
        "packetId": packet_id,
        "archiveSha256": sha256_file(archive_path),
        "payloadFiles": len(expected_payload_paths),
        "authorityEffect": AUTHORITY_EFFECT,
    }


def command_ingest_response(args: argparse.Namespace) -> dict[str, Any]:
    source = Path(args.response).resolve()
    entries = read_response_source(source)
    required = {"INPUT_PACKET_ID.txt", "response.md"}
    missing = required - entries.keys()
    if missing:
        raise BridgeError("MISSING_REQUIRED:" + ",".join(sorted(missing)))
    packet_id = entries["INPUT_PACKET_ID.txt"].decode("utf-8").strip().strip("`")
    if packet_id != args.expected_packet_id:
        raise BridgeError("RESPONSE_PACKET_BINDING_MISMATCH")

    for name, data in entries.items():
        validate_payload_name(name)
        if name not in required and not name.startswith("artifacts/"):
            raise BridgeError(f"RESPONSE_PATH_OUTSIDE_CONTRACT:{name}")
        scan_for_secrets(name, data)

    file_records = [
        {
            "path": name,
            "bytes": len(data),
            "sha256": sha256_bytes(data),
        }
        for name, data in sorted(entries.items())
    ]
    response_core = {
        "schema": RECEIPT_SCHEMA,
        "schemaVersion": SCHEMA_VERSION,
        "canonicalizationProfile": CANONICAL_PROFILE,
        "inputPacketId": packet_id,
        "source": "GROK_REMOTE_SANDBOX_DECLARED",
        "declaredRemoteDropLocation": REMOTE_DROP,
        "eventType": "EXTERNAL_RESPONSE_STAGED",
        "status": "QUARANTINED_UNTRUSTED",
        "governanceState": GOVERNANCE_STATE,
        "authorityEffect": AUTHORITY_EFFECT,
        "files": file_records,
        "epistemicLimits": [
            "Receipt proves local bytes and declared packet binding only.",
            "Receipt does not prove responder identity, semantic truth, lawful ownership, or completeness.",
            "Receipt grants no Owner, Seal, Apply, or state-commit authority.",
        ],
    }
    response_root = sha256_bytes(canonical_bytes(response_core))
    receipt = dict(response_core)
    receipt["responseRoot"] = response_root

    stage_base = Path(args.stage_dir).resolve()
    stage_target = stage_base / response_root.removeprefix("sha256:")[:24]
    if stage_target.exists():
        raise BridgeError(f"STAGE_TARGET_ALREADY_EXISTS:{stage_target}")
    stage_target.mkdir(parents=True)
    try:
        for name, data in entries.items():
            target = stage_target / Path(*PurePosixPath(name).parts)
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_bytes(data)
        (stage_target / "GROK_RESPONSE_INTAKE_RECEIPT.json").write_bytes(
            canonical_bytes(receipt) + b"\n"
        )
    except Exception:
        shutil.rmtree(stage_target, ignore_errors=True)
        raise

    return {
        "status": "RESPONSE_STAGED",
        "inputPacketId": packet_id,
        "responseRoot": response_root,
        "stagePath": str(stage_target),
        "authorityEffect": AUTHORITY_EFFECT,
        "disposition": "QUARANTINED_UNTRUSTED",
    }


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description=__doc__)
    subparsers = parser.add_subparsers(dest="command", required=True)

    pack = subparsers.add_parser("pack", help="Build a deterministic Grok review packet")
    pack.add_argument("--task", required=True, help="Markdown task file")
    pack.add_argument("--include", action="append", default=[], help="Explicit file to include")
    pack.add_argument("--output", required=True, help="Output ZIP path")
    pack.set_defaults(handler=command_pack)

    verify = subparsers.add_parser("verify-export", help="Verify an outgoing packet")
    verify.add_argument("archive", help="Packet ZIP")
    verify.set_defaults(handler=command_verify_export)

    ingest = subparsers.add_parser(
        "ingest-response", help="Verify and quarantine a Grok response folder or ZIP"
    )
    ingest.add_argument("--response", required=True, help="Response folder or ZIP")
    ingest.add_argument("--expected-packet-id", required=True)
    ingest.add_argument("--stage-dir", required=True)
    ingest.set_defaults(handler=command_ingest_response)
    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    try:
        result = args.handler(args)
        print(json.dumps(result, indent=2, sort_keys=True))
        return 0
    except (BridgeError, OSError, ValueError, zipfile.BadZipFile, json.JSONDecodeError) as error:
        print(
            json.dumps(
                {
                    "status": "REJECTED",
                    "reason": str(error),
                    "authorityEffect": AUTHORITY_EFFECT,
                },
                indent=2,
                sort_keys=True,
            ),
            file=sys.stderr,
        )
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
