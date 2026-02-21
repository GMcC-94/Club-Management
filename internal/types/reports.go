package types

// FinancialStats for financial reports
type FinancialStats struct {
	TotalIncome           float64
	TotalExpenditure      float64
	NetBalance            float64
	IncomeByCategory      []ChartDataItem
	ExpenditureByCategory []ChartDataItem
	MonthlyTrend          []FinancialTrendItem
}

// ChartDataItem for simple charts
type ChartDataItem struct {
	Label string
	Value int
}

// TrendDataItem for trend charts
type TrendDataItem struct {
	Date  string
	Value int
}

// FinancialTrendItem for financial trend charts
type FinancialTrendItem struct {
	Date        string
	Income      float64
	Expenditure float64
}
