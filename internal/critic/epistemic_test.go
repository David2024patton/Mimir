package critic

import (
	"testing"
)

func TestZeroFabricationOfFlawsWhenGrounded(t *testing.T) {
	ctx := "Mímir passed 11 Go package test suites in 8.1 seconds and saved 20 neurons."
	draft := "Mímir passed 11 Go package test suites and saved 20 neurons."

	rep := AuditDeterministic(ctx, draft)
	if !rep.PassAudit {
		t.Fatalf("expected PassAudit=true for grounded report, got false (%s)", rep.RoutingReason)
	}
	if len(rep.UnbackedClaims) != 0 || len(rep.VendorBiasFlags) != 0 || len(rep.HypeFlags) != 0 {
		t.Errorf("expected empty violation arrays, got %+v", rep)
	}
}

func TestFlagsUnbackedMetricsVendorBiasAndHype(t *testing.T) {
	ctx := "WordStream vendor benchmark reports $90.92 CPL across campaigns."
	draft := "Our engine is guaranteed to be 10x faster with <0.2ms latency, while the WordStream vendor benchmark reports $90.92 CPL."

	rep := AuditDeterministic(ctx, draft)
	if rep.PassAudit {
		t.Fatal("expected PassAudit=false when HIGH severity unbacked metrics (<0.2ms, 10x faster) are present")
	}
	if len(rep.UnbackedClaims) == 0 {
		t.Error("expected UnbackedClaims to flag <0.2ms and 10x faster")
	}
	if len(rep.VendorBiasFlags) == 0 {
		t.Error("expected VendorBiasFlags to flag $90.92 vendor benchmark")
	}
	if len(rep.HypeFlags) < 2 {
		t.Errorf("expected HypeFlags to flag 'guaranteed' and '10x faster', got %+v", rep.HypeFlags)
	}
}

func TestRoutingLogicMoreThanTwoLowTakeawaysFails(t *testing.T) {
	rep := AuditReport{
		CorrectedTakeaways: []AuditedTakeaway{
			{Index: 1, CorrectedTakeaway: "T1", Confidence: ConfidenceLow},
			{Index: 2, CorrectedTakeaway: "T2", Confidence: ConfidenceLow},
			{Index: 3, CorrectedTakeaway: "T3", Confidence: ConfidenceLow},
		},
	}
	rep.ApplyRoutingRule()
	if rep.PassAudit {
		t.Fatal("expected PassAudit=false when >2 takeaways have LOW confidence")
	}
}
