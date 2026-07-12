package ai

import (
	"context"
	"testing"
)

func TestFixtureReviewEvidence(t *testing.T) {
	f := Fixture{}
	ctx := context.Background()

	ok, err := f.ReviewEvidence(ctx, "https://cdn.example/proof-tree-planting.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if !ok.LooksValid || ok.Confidence < 0.5 {
		t.Fatalf("expected valid proof fixture, got %+v", ok)
	}

	missing, err := f.ReviewEvidence(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	if missing.LooksValid {
		t.Fatalf("empty URL should not look valid: %+v", missing)
	}

	blur, err := f.ReviewEvidence(ctx, "https://cdn.example/blur-unclear.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if blur.LooksValid {
		t.Fatalf("blurry proof should flag needs-review: %+v", blur)
	}

	up, err := f.Review(EvidenceInput{DataURL: "data:image/png;base64,aaa", FileName: "tree.jpg"})
	if err != nil || !up.LooksValid {
		t.Fatalf("uploaded data url fixture: %+v %v", up, err)
	}
}

func TestGatewayUsesFixtureWhenNoKey(t *testing.T) {
	g := NewGateway("", "openai/gpt-4.1-mini", false)
	if !g.UseFixture() {
		t.Fatal("expected fixture mode when API key empty")
	}
	rev, err := g.ReviewEvidence(context.Background(), "https://cdn.example/ok.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if !rev.LooksValid {
		t.Fatalf("fixture path failed: %+v", rev)
	}
}

func TestFixtureNarrativeIsAdvisory(t *testing.T) {
	s, err := Fixture{}.Summarize(context.Background(), map[string]any{"overall": 72})
	if err != nil {
		t.Fatal(err)
	}
	if s == "" {
		t.Fatal("empty narrative")
	}
}
