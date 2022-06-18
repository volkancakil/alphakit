// Copyright 2022 The Coln Group Ltd
// SPDX-License-Identifier: MIT

package perf

import (
	"time"

	"github.com/thecolngroup/alphakit/broker"
	"github.com/thecolngroup/gou/num"
)

// PortfolioReport is report on the portfolio metrics.
// It is generated by the NewPortfolioReport function using an equity curve,
// typically the output of algo execution using a Dealer in package broker.
type PortfolioReport struct {

	// PeriodStart is the start time of the equity curve.
	PeriodStart time.Time

	// PeriodEnd is the end time of the equity curve.
	PeriodEnd time.Time

	// Period is the duration of the equity curve.
	Period time.Duration

	// StartEquity is the starting equity (amount) of the equity curve.
	StartEquity float64

	// EndEquity is the ending equity (amount) of the equity curve.
	EndEquity float64

	// EquityReturn is the percentage return of the equity curve.
	EquityReturn float64

	// CAGR is the Compound Annual Growth Rate of the equity curve.
	CAGR float64

	// MaxDrawdown is the maximum percentage drawdown of the equity curve
	MaxDrawdown float64

	// MDDRecovery is the recovery time of the maximum drawdown of the equity curve.
	MDDRecovery time.Duration

	// HistVolAnn is the historic volatility of the equity curve as annualized std dev.
	HistVolAnn float64

	// Sharpe is the Sharpe ratio of the equity curve.
	Sharpe float64

	// Calmar is the Calmar ratio of the equity curve.
	Calmar float64

	// EquityCurve is the source from which the report fields are generated.
	EquityCurve broker.EquitySeries `csv:"-"`

	drawdowns []Drawdown
	mdd       Drawdown
}

// NewPortfolioReport creates a new PortfolioReport from a given equity curve.
func NewPortfolioReport(curve broker.EquitySeries) *PortfolioReport {
	if len(curve) == 0 {
		return nil
	}

	t := curve.SortKeys()
	tStart, tEnd := t[0], t[len(t)-1]

	var report PortfolioReport
	report.EquityCurve = curve
	report.PeriodStart, report.StartEquity = tStart.Time(), curve[tStart].InexactFloat64()
	report.PeriodEnd, report.EndEquity = tEnd.Time(), curve[tEnd].InexactFloat64()
	report.Period = report.PeriodEnd.Sub(report.PeriodStart)

	report.EquityReturn = (report.EndEquity - report.StartEquity) / num.NNZ(report.StartEquity, 1)
	report.CAGR = num.NN(CAGR(report.StartEquity, report.EndEquity, int(report.Period.Hours())/24), 0)

	daily := ReduceEOD(curve)
	if len(daily) == 0 {
		return &report
	}

	report.drawdowns = Drawdowns(daily)
	report.mdd = MaxDrawdown(report.drawdowns)
	report.MaxDrawdown = report.mdd.Pct
	report.MDDRecovery = report.mdd.Recovery

	returns := DiffPctReturns(daily)

	report.HistVolAnn = HistVolAnn(returns)
	report.Sharpe = SharpeRatio(returns, SharpeDefaultAnnualRiskFreeRate)
	report.Calmar = CalmarRatio(report.CAGR, report.MaxDrawdown)

	return &report
}
