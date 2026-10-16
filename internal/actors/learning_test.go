package actors

import (
	"encoding/json"
	"github.com/caglarkc/trading/contracts"
	"github.com/caglarkc/trading/internal/domain"
	"github.com/caglarkc/trading/internal/store"
	"testing"
)

func lessonFixture() LearningReport {
	return LearningReport{Summary: "Sınırlı işlem örneğinde süreç hatası ve normal kayıp ayrılacak.", OpenReview: "Açık pozisyon için henüz kapanmış sonuç değerlendirmesi yok.",
		Reviews: []TradeLesson{{FillID: "own-fill", Assessment: "Beklenen toparlanma gerçekleşmedi; bu tek örnek yöntem başarısını ölçmez.", Classification: "insufficient_evidence", Lesson: "Yeniden girişte gözlenebilir fiyat teyidi denenecek.", RuleIDs: []string{"entry-v1"}}},
		Rules:   []StrategyRule{{ID: "entry-v1", Previous: "Önceki giriş koşulu kayıtta eksik.", Current: "Yeniden giriş öncesi belirlenen fiyat koşulunu doğrula.", Reason: "Tek kayıptan başarı çıkarma; girişteki eksik kanıtı açık tut.", Status: "trial", Evidence: []string{"own-fill"}, TestPlan: "Sonraki kapanan işlemlerde net sonuç ve kural uyumunu karşılaştır.", Followup: "Henüz sonraki işlem yok; etkinlik test edilmedi."}}}
}
func TestLearningCoverageScopeAndRuleLifecycle(t *testing.T) {
	f := []store.LearningFill{{ID: "own-fill", Seq: 2}}
	empty := store.StrategyState{Report: json.RawMessage(`{}`)}
	valid := lessonFixture()
	if err := contracts.Validate(LearningSchema(), bytesOf(valid)); err != nil {
		t.Fatal(err)
	}
	if err := ValidateLearning(valid, empty, f); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		change func(*LearningReport)
	}{
		{"missing_closed_fill", func(r *LearningReport) { r.Reviews = nil }},
		{"duplicate_closed_fill", func(r *LearningReport) { r.Reviews = append(r.Reviews, r.Reviews[0]) }},
		{"foreign_fill", func(r *LearningReport) { r.Reviews[0].FillID = "peer-fill" }},
		{"foreign_rule_evidence", func(r *LearningReport) { r.Rules[0].Evidence = []string{"peer-fill"} }},
		{"unrecorded_rule", func(r *LearningReport) { r.Reviews[0].RuleIDs = []string{"invented"} }},
		{"unproven_new_rule", func(r *LearningReport) { r.Rules[0].Status = "retained" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := lessonFixture()
			tc.change(&r)
			if ValidateLearning(r, empty, f) == nil {
				t.Fatal("invalid report accepted")
			}
		})
	}
	old := store.StrategyState{Version: 1, Through: 2, Report: bytesOf(valid)}
	r := lessonFixture()
	r.Reviews = nil
	r.Rules = nil
	if ValidateLearning(r, old, nil) == nil {
		t.Fatal("silent deletion of previous rule accepted")
	}
	r = lessonFixture()
	r.Reviews = nil
	r.Rules[0].Current = "Different entry condition that is now being tested."
	if ValidateLearning(r, old, nil) == nil {
		t.Fatal("untracked strategy change accepted")
	}
	r.Rules[0].Previous = valid.Rules[0].Current
	if err := ValidateLearning(r, old, nil); err != nil {
		t.Fatal(err)
	}
	// No historical exits is a valid baseline, not invented evidence.
	r = lessonFixture()
	r.Reviews = nil
	r.Rules[0].Evidence = nil
	if err := ValidateLearning(r, empty, nil); err != nil {
		t.Fatal(err)
	}
}
func TestDecisionMustReferenceItsOwnStrategySnapshot(t *testing.T) {
	state := store.StrategyState{Version: 2, Report: bytesOf(lessonFixture())}
	a := domain.LearningApplication{Version: 2, RuleIDs: []string{"entry-v1"}, Explanation: "This decision waits for the recorded confirmation before buying."}
	if err := ValidateLearningApplication(a, state); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []domain.LearningApplication{
		{Version: 1, RuleIDs: a.RuleIDs}, {Version: 2, RuleIDs: []string{"peer-rule"}}, {Version: 2}, {Version: 2, RuleIDs: []string{"entry-v1", "entry-v1"}},
	} {
		if ValidateLearningApplication(bad, state) == nil {
			t.Fatal("invalid strategy application accepted", bad)
		}
	}
}
