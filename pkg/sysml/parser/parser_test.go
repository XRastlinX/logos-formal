package parser

import (
	"os"
	"reflect"
	"testing"

	"github.com/XRastlinX/logos-formal/pkg/sysml/ast"
)

func TestParser_1_InputParametersAndStrictCalcReturn(t *testing.T) {
	result := ParsePackage([]byte(`
package Pilot {
    calc def Scale {
        in attribute input :> ISQ::length;
        in attribute reference :> ISQ::length;
        return :> ISQ::dimensionless = input / reference;
    }
}`))

	requireParsed(t, result)
	if len(result.Package.Calcs) != 1 {
		t.Fatalf("expected 1 calc, got %d", len(result.Package.Calcs))
	}
	calc := result.Package.Calcs[0]
	if calc.Name != "Scale" {
		t.Fatalf("expected calc name Scale, got %q", calc.Name)
	}
	if len(calc.Inputs) != 2 {
		t.Fatalf("expected 2 ordered inputs, got %d", len(calc.Inputs))
	}
	if calc.Inputs[0].Name != "input" || calc.Inputs[1].Name != "reference" {
		t.Fatalf("input order was not preserved: %+v", calc.Inputs)
	}
	if calc.Return.Quantity != "ISQ::dimensionless" {
		t.Fatalf("unexpected return quantity: %q", calc.Return.Quantity)
	}
	if _, ok := calc.Body.(*ast.InfixExpression); !ok {
		t.Fatalf("expected infix body, got %T", calc.Body)
	}
}

func TestParser_2_ConstraintBareBooleanExpression(t *testing.T) {
	result := ParsePackage([]byte(`
package Pilot {
    constraint def Domain {
        in attribute altitude :> ISQ::length;
        altitude >= 0[km] and altitude <= 5[km]
    }
}`))

	requireParsed(t, result)
	if len(result.Package.Constraints) != 1 {
		t.Fatalf("expected 1 constraint, got %d", len(result.Package.Constraints))
	}
	constraint := result.Package.Constraints[0]
	if constraint.Name != "Domain" || len(constraint.Inputs) != 1 {
		t.Fatalf("unexpected constraint: %+v", constraint)
	}
	root, ok := constraint.Body.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expected infix constraint root, got %T", constraint.Body)
	}
	if root.Operator != "and" {
		t.Fatalf("expected boolean conjunction at root, got %q", root.Operator)
	}
}

func TestParser_3_OperatorPrecedence(t *testing.T) {
	result := ParsePackage([]byte(`
package Pilot {
    calc def Formula {
        in attribute a :> ISQ::length;
        in attribute b :> ISQ::length;
        return :> ISQ::length = 1 + a * b - -a;
    }
}`))

	requireParsed(t, result)
	body := result.Package.Calcs[0].Body

	minus, ok := body.(*ast.InfixExpression)
	if !ok || minus.Operator != "-" {
		t.Fatalf("expected subtraction root, got %#v", body)
	}
	plus, ok := minus.Left.(*ast.InfixExpression)
	if !ok || plus.Operator != "+" {
		t.Fatalf("expected addition on left, got %#v", minus.Left)
	}
	product, ok := plus.Right.(*ast.InfixExpression)
	if !ok || product.Operator != "*" {
		t.Fatalf("expected multiplication below addition, got %#v", plus.Right)
	}
	prefix, ok := minus.Right.(*ast.PrefixExpression)
	if !ok || prefix.Operator != "-" {
		t.Fatalf("expected prefix minus on right, got %#v", minus.Right)
	}
}

func TestParser_4_NestedCalculationCalls(t *testing.T) {
	result := ParsePackage([]byte(`
package Pilot {
    calc def Composite {
        in attribute altitude :> ISQ::length;
        in attribute reference :> ISQ::pressure;
        return :> ISQ::force =
            PressureToDrag(AltitudeToPressure(altitude, reference), reference);
    }
}`))

	requireParsed(t, result)
	outer, ok := result.Package.Calcs[0].Body.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected outer call, got %T", result.Package.Calcs[0].Body)
	}
	if len(outer.Args) != 2 {
		t.Fatalf("expected 2 outer args, got %d", len(outer.Args))
	}
	inner, ok := outer.Args[0].(*ast.CallExpression)
	if !ok || len(inner.Args) != 2 {
		t.Fatalf("expected nested two-argument call, got %#v", outer.Args[0])
	}
}

func TestParser_5_StrictMemberOrderingIsNoDecision(t *testing.T) {
	inputs := []string{
		`
package Pilot {
    calc def Reordered {
        return :> ISQ::length = x;
        in attribute x :> ISQ::length;
    }
}`,
		`
package Pilot {
    constraint def Reordered {
        x >= 0[km]
        in attribute x :> ISQ::length;
    }
}`,
	}

	for _, input := range inputs {
		result := ParsePackage([]byte(input))
		if result.Status != StatusNoDecision {
			t.Fatalf("expected NO_DECISION for unsupported ordering, got %s: %+v", result.Status, result.Diagnostics)
		}
		if len(result.Diagnostics) != 1 {
			t.Fatalf("expected one diagnostic, got %+v", result.Diagnostics)
		}
	}
}

func TestParser_6_MalformedSupportedSyntaxIsReject(t *testing.T) {
	tests := []struct {
		name string
		src  string
		code string
	}{
		{
			name: "missing input semicolon",
			src: `
package Pilot {
    calc def Broken {
        in attribute x :> ISQ::length
        return :> ISQ::length = x;
    }
}`,
			code: "SYNTAX_ERROR",
		},
		{
			name: "missing return",
			src: `
package Pilot {
    calc def Broken {
        in attribute x :> ISQ::length;
    }
}`,
			code: "MISSING_RETURN",
		},
		{
			name: "lexical error",
			src: `
package Pilot {
    calc def Broken {
        return :> ISQ::length = 1.2.3;
    }
}`,
			code: "LEXICAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParsePackage([]byte(tt.src))
			if result.Status != StatusReject {
				t.Fatalf("expected REJECT, got %s: %+v", result.Status, result.Diagnostics)
			}
			if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != tt.code {
				t.Fatalf("expected diagnostic %s, got %+v", tt.code, result.Diagnostics)
			}
			if result.Package != nil {
				t.Fatal("rejected parse must not expose a package")
			}
		})
	}
}

func TestParser_7_UnsupportedFormsAreNoDecision(t *testing.T) {
	tests := []string{
		`package Pilot { private import ISQ::Quantities; }`,
		`
package Pilot {
    constraint def Domain {
        in attribute x :> ISQ::length;
        x >= 0[km];
    }
}`,
		`package Pilot {} package Other {}`,
		`
package Pilot {
    calc def UnsupportedCallTarget {
        in attribute x :> ISQ::length;
        in attribute y :> ISQ::length;
        return :> ISQ::length = (x + y)(x);
    }
}`,
	}

	for _, input := range tests {
		result := ParsePackage([]byte(input))
		if result.Status != StatusNoDecision {
			t.Fatalf("expected NO_DECISION, got %s: %+v", result.Status, result.Diagnostics)
		}
		if result.Package != nil {
			t.Fatal("NO_DECISION parse must not expose a package")
		}
	}
}

func TestParser_8_GoldenSysMLFixture(t *testing.T) {
	source, err := os.ReadFile("../../../04_Dual_Projection/CTPhase0ScalarChain.sysml")
	if err != nil {
		t.Fatalf("read golden fixture: %v", err)
	}

	result := ParsePackage(source)
	requireParsed(t, result)
	if result.Package.Name != "CTPhase0ScalarChain" {
		t.Fatalf("unexpected package name: %q", result.Package.Name)
	}
	if len(result.Package.Calcs) != 3 {
		t.Fatalf("expected exactly 3 calcs, got %d", len(result.Package.Calcs))
	}
	if len(result.Package.Constraints) != 3 {
		t.Fatalf("expected exactly 3 constraints, got %d", len(result.Package.Constraints))
	}

	gotCalcs := []string{
		result.Package.Calcs[0].Name,
		result.Package.Calcs[1].Name,
		result.Package.Calcs[2].Name,
	}
	wantCalcs := []string{"AltitudeToPressure", "PressureToDrag", "AltitudeToDrag"}
	if !reflect.DeepEqual(gotCalcs, wantCalcs) {
		t.Fatalf("unexpected calc order: got %v, want %v", gotCalcs, wantCalcs)
	}

	gotConstraints := []string{
		result.Package.Constraints[0].Name,
		result.Package.Constraints[1].Name,
		result.Package.Constraints[2].Name,
	}
	wantConstraints := []string{"PilotAltitudeDomain", "PilotPressureDomain", "PilotDragDomain"}
	if !reflect.DeepEqual(gotConstraints, wantConstraints) {
		t.Fatalf("unexpected constraint order: got %v, want %v", gotConstraints, wantConstraints)
	}

	packageSpan := result.Package.Span()
	if packageSpan.StartByte < 0 || packageSpan.EndByte > len(source) ||
		packageSpan.StartByte >= packageSpan.EndByte {
		t.Fatalf("invalid package byte span: %+v for %d bytes", packageSpan, len(source))
	}
	if got := string(source[packageSpan.StartByte:packageSpan.EndByte]); got[0:7] != "package" || got[len(got)-1:] != "}" {
		t.Fatalf("package span does not bind the exact declaration: %q", got)
	}

	bodySpan := result.Package.Calcs[0].Body.Span()
	if bodySpan.StartByte < packageSpan.StartByte || bodySpan.EndByte > packageSpan.EndByte {
		t.Fatalf("calc body span escapes package span: body=%+v package=%+v", bodySpan, packageSpan)
	}
}

func TestParser_9_DeterministicRepeatedParse(t *testing.T) {
	source := []byte(`
package Pilot {
    calc def Identity {
        in attribute x :> ISQ::length;
        return :> ISQ::length = x;
    }
}`)

	first := ParsePackage(source)
	second := ParsePackage(source)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated parses diverged:\nfirst:  %#v\nsecond: %#v", first, second)
	}
}

func requireParsed(t *testing.T, result Result) {
	t.Helper()
	if result.Status != StatusParsed {
		t.Fatalf("expected PARSED, got %s: %+v", result.Status, result.Diagnostics)
	}
	if result.Package == nil {
		t.Fatal("parsed result has nil package")
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("parsed result has diagnostics: %+v", result.Diagnostics)
	}
}
