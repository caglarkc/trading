package actors

import (
	"fmt"
	"testing"
)

func TestReviewCoverageAndCapturedEvidence(t *testing.T) {
	data := ReviewDataset{}
	report := ReviewReport{MarketOutlook: "The broad market review compares technology, financials, energy, consumer businesses and prevailing macro conditions using captured evidence."}
	for i := 0; i < 50; i++ {
		symbol := fmt.Sprintf("S%d", i)
		id := fmt.Sprintf("e%d", i)
		text := fmt.Sprintf("Company %s recorded revenue growth and trades at a published valuation with business-specific operating risks.", symbol)
		data.Documents = append(data.Documents, ReviewDocument{Symbol: symbol, EvidenceID: id, Text: text})
		report.Evaluations = append(report.Evaluations, Evaluation{Symbol: symbol, EvidenceID: id, Excerpt: text, Assessment: fmt.Sprintf("The opportunity for %s is evaluated against peers and its own valuation and business outlook.", symbol), Risk: "Company operating margins could weaken.", Catalyst: "Upcoming earnings publication.", Score: 50})
	}
	for i := 0; i < 20; i++ {
		report.Watchlist = append(report.Watchlist, WatchCandidate{Symbol: fmt.Sprintf("X%d", i), Exchange: "NYSE", Reason: "This company adds a specific differentiated sector and catalyst to the watchlist."})
	}
	if e := ValidateReview(report, data, true, "outside"); e != nil {
		t.Fatal(e)
	}
	original := report.Evaluations[0]
	report.Evaluations[0].Excerpt = "This supporting statement was never in the captured page."
	if ValidateReview(report, data, true, "outside") == nil {
		t.Fatal("fabricated source excerpt accepted")
	}
	report.Evaluations[0] = original
	report.Evaluations[0].Symbol = report.Evaluations[1].Symbol
	if ValidateReview(report, data, true, "outside") == nil {
		t.Fatal("duplicate stock hid missing coverage")
	}
	report.Evaluations[0] = original
	report.Watchlist[0].Symbol = data.Documents[0].Symbol
	if ValidateReview(report, data, true, "outside") == nil {
		t.Fatal("fixed stock accepted as outside candidate")
	}
}
