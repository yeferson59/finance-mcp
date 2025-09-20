# Usage Examples for Simple MCP Market Data Server

This document provides practical code examples for integrating the Simple MCP Market Data Server with AI applications, chatbots, and automated financial analysis systems.

## Table of Contents

1. [Basic MCP Client Setup](#basic-mcp-client-setup)
2. [Financial Analysis Assistant](#financial-analysis-assistant)
3. [Investment Research Bot](#investment-research-bot)
4. [Portfolio Screening Tool](#portfolio-screening-tool)
5. [Risk Assessment System](#risk-assessment-system)
6. [Educational Finance Chatbot](#educational-finance-chatbot)
7. [Market Comparison Tool](#market-comparison-tool)
8. [Integration with Popular AI Frameworks](#integration-with-popular-ai-frameworks)

## Basic MCP Client Setup

### Python MCP Client Example

```python
import asyncio
import json
import subprocess
import sys
from typing import Dict, Any, Optional

class SimpleMCPClient:
    """Basic MCP client for connecting to Simple MCP Market Data Server"""

    def __init__(self, server_path: str):
        self.server_path = server_path
        self.process = None

    async def start(self):
        """Start the MCP server process"""
        self.process = await asyncio.create_subprocess_exec(
            self.server_path,
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE
        )

        # Initialize MCP connection
        await self._send_request({
            "jsonrpc": "2.0",
            "id": "init",
            "method": "initialize",
            "params": {
                "protocolVersion": "2024-11-05",
                "capabilities": {}
            }
        })

    async def get_stock_data(self, symbol: str) -> Dict[str, Any]:
        """Fetch stock data for given symbol"""
        response = await self._send_request({
            "jsonrpc": "2.0",
            "id": f"get-stock-{symbol}",
            "method": "tools/call",
            "params": {
                "name": "get-stock",
                "arguments": {"symbol": symbol.upper()}
            }
        })

        if "error" in response:
            raise Exception(f"Error fetching {symbol}: {response['error']['message']}")

        # Parse the JSON content from the response
        content = response["result"]["content"][0]["text"]
        return json.loads(content)

    async def _send_request(self, request: Dict) -> Dict:
        """Send JSON-RPC request to MCP server"""
        if not self.process:
            raise Exception("Server not started")

        request_json = json.dumps(request) + "\n"
        self.process.stdin.write(request_json.encode())
        await self.process.stdin.drain()

        response_line = await self.process.stdout.readline()
        return json.loads(response_line.decode().strip())

    async def close(self):
        """Close the MCP server connection"""
        if self.process:
            self.process.terminate()
            await self.process.wait()

# Usage example
async def main():
    client = SimpleMCPClient("./simple-mcp")

    try:
        await client.start()

        # Fetch Apple stock data
        apple_data = await client.get_stock_data("AAPL")
        print(f"Apple Inc: ${apple_data.get('MarketCapitalization', 'N/A')} market cap")

    finally:
        await client.close()

if __name__ == "__main__":
    asyncio.run(main())
```

## Financial Analysis Assistant

### AI-Powered Stock Analyzer

```python
import asyncio
from typing import List, Dict, Tuple
from dataclasses import dataclass

@dataclass
class StockAnalysis:
    symbol: str
    name: str
    recommendation: str
    key_metrics: Dict[str, float]
    risks: List[str]
    opportunities: List[str]
    summary: str

class FinancialAnalysisAssistant:
    """AI-powered financial analysis using MCP stock data"""

    def __init__(self, mcp_client):
        self.client = mcp_client

    async def analyze_stock(self, symbol: str) -> StockAnalysis:
        """Comprehensive stock analysis"""
        try:
            data = await self.client.get_stock_data(symbol)

            # Extract key metrics
            metrics = self._extract_metrics(data)

            # Perform analysis
            recommendation = self._generate_recommendation(metrics, data)
            risks = self._identify_risks(metrics, data)
            opportunities = self._identify_opportunities(metrics, data)
            summary = self._generate_summary(data, metrics, recommendation)

            return StockAnalysis(
                symbol=symbol,
                name=data.get("Name", "Unknown"),
                recommendation=recommendation,
                key_metrics=metrics,
                risks=risks,
                opportunities=opportunities,
                summary=summary
            )

        except Exception as e:
            raise Exception(f"Analysis failed for {symbol}: {str(e)}")

    def _extract_metrics(self, data: Dict) -> Dict[str, float]:
        """Extract and convert key financial metrics"""
        metrics = {}

        # Safe conversion function
        def safe_float(value, default=0.0):
            try:
                return float(value) if value else default
            except (ValueError, TypeError):
                return default

        metrics["pe_ratio"] = safe_float(data.get("PERatio"))
        metrics["market_cap"] = safe_float(data.get("MarketCapitalization"))
        metrics["dividend_yield"] = safe_float(data.get("DividendYield"))
        metrics["profit_margin"] = safe_float(data.get("ProfitMargin"))
        metrics["roe"] = safe_float(data.get("ReturnOnEquityTTM"))
        metrics["beta"] = safe_float(data.get("Beta"))
        metrics["eps"] = safe_float(data.get("EPS"))
        metrics["book_value"] = safe_float(data.get("BookValue"))

        return metrics

    def _generate_recommendation(self, metrics: Dict, data: Dict) -> str:
        """Generate investment recommendation based on metrics"""
        score = 0

        # P/E ratio analysis
        pe_ratio = metrics["pe_ratio"]
        if 0 < pe_ratio < 15:
            score += 2  # Undervalued
        elif 15 <= pe_ratio < 25:
            score += 1  # Fair value
        elif pe_ratio >= 30:
            score -= 1  # Overvalued

        # Profitability analysis
        profit_margin = metrics["profit_margin"]
        if profit_margin > 0.2:  # 20%+
            score += 2
        elif profit_margin > 0.1:  # 10%+
            score += 1
        elif profit_margin < 0:
            score -= 2

        # ROE analysis
        roe = metrics["roe"]
        if roe > 0.15:  # 15%+
            score += 1
        elif roe < 0:
            score -= 1

        # Dividend analysis
        div_yield = metrics["dividend_yield"]
        if div_yield > 0.03:  # 3%+
            score += 1

        # Generate recommendation
        if score >= 4:
            return "Strong Buy"
        elif score >= 2:
            return "Buy"
        elif score >= 0:
            return "Hold"
        elif score >= -2:
            return "Weak Hold"
        else:
            return "Sell"

    def _identify_risks(self, metrics: Dict, data: Dict) -> List[str]:
        """Identify potential investment risks"""
        risks = []

        # Valuation risk
        if metrics["pe_ratio"] > 40:
            risks.append("High valuation - P/E ratio above 40x suggests overvaluation")

        # Volatility risk
        if metrics["beta"] > 2.0:
            risks.append("High volatility - Beta above 2.0 indicates significant price swings")

        # Profitability risk
        if metrics["profit_margin"] < 0.05:  # Less than 5%
            risks.append("Low profitability - Profit margins below 5% indicate weak earnings")

        # Dividend sustainability
        if metrics["dividend_yield"] > 0.08:  # Above 8%
            risks.append("High dividend yield may indicate dividend sustainability concerns")

        # Market cap risk
        if metrics["market_cap"] < 1e9:  # Less than $1B
            risks.append("Small market cap may result in higher volatility and liquidity risk")

        return risks

    def _identify_opportunities(self, metrics: Dict, data: Dict) -> List[str]:
        """Identify potential opportunities"""
        opportunities = []

        # Value opportunity
        if 0 < metrics["pe_ratio"] < 12:
            opportunities.append("Value opportunity - Low P/E ratio suggests undervaluation")

        # High profitability
        if metrics["profit_margin"] > 0.25:  # Above 25%
            opportunities.append("High profitability - Strong profit margins indicate efficient operations")

        # Strong ROE
        if metrics["roe"] > 0.20:  # Above 20%
            opportunities.append("Strong returns - High ROE indicates effective capital utilization")

        # Dividend income
        if 0.02 < metrics["dividend_yield"] < 0.06:  # 2-6% range
            opportunities.append("Steady dividend income with sustainable yield")

        # Growth potential
        sector = data.get("Sector", "").lower()
        if sector in ["technology", "healthcare", "renewable energy"]:
            opportunities.append(f"Growth sector exposure - {sector} sector has long-term growth potential")

        return opportunities

    def _generate_summary(self, data: Dict, metrics: Dict, recommendation: str) -> str:
        """Generate analysis summary"""
        name = data.get("Name", "Unknown Company")
        sector = data.get("Sector", "Unknown")

        summary = f"{name} ({data.get('Symbol', 'N/A')}) operates in the {sector} sector. "

        # Add key metric highlights
        if metrics["market_cap"] > 1e12:  # $1T+
            summary += "This large-cap stock "
        elif metrics["market_cap"] > 1e10:  # $10B+
            summary += "This mid-cap stock "
        else:
            summary += "This small-cap stock "

        summary += f"has a P/E ratio of {metrics['pe_ratio']:.1f} and profit margins of {metrics['profit_margin']*100:.1f}%. "

        # Add recommendation context
        summary += f"Based on fundamental analysis, the recommendation is '{recommendation}'. "

        # Add key insight
        if metrics["dividend_yield"] > 0:
            summary += f"The stock pays a {metrics['dividend_yield']*100:.1f}% dividend yield."

        return summary

# Usage example
async def analyze_portfolio():
    client = SimpleMCPClient("./simple-mcp")
    analyzer = FinancialAnalysisAssistant(client)

    try:
        await client.start()

        symbols = ["AAPL", "GOOGL", "MSFT", "AMZN"]

        for symbol in symbols:
            try:
                analysis = await analyzer.analyze_stock(symbol)

                print(f"\n=== {analysis.name} ({analysis.symbol}) ===")
                print(f"Recommendation: {analysis.recommendation}")
                print(f"Summary: {analysis.summary}")

                if analysis.risks:
                    print("Key Risks:")
                    for risk in analysis.risks:
                        print(f"  • {risk}")

                if analysis.opportunities:
                    print("Opportunities:")
                    for opp in analysis.opportunities:
                        print(f"  • {opp}")

            except Exception as e:
                print(f"Failed to analyze {symbol}: {e}")

    finally:
        await client.close()

if __name__ == "__main__":
    asyncio.run(analyze_portfolio())
```

## Investment Research Bot

### Automated Research Report Generator

```python
import asyncio
from datetime import datetime
from typing import Dict, List
import json

class InvestmentResearchBot:
    """Automated investment research report generation"""

    def __init__(self, mcp_client):
        self.client = mcp_client

    async def generate_research_report(self, symbol: str) -> str:
        """Generate comprehensive research report"""
        try:
            data = await self.client.get_stock_data(symbol)

            # Generate report sections
            report = self._generate_report_header(data)
            report += self._generate_company_overview(data)
            report += self._generate_financial_analysis(data)
            report += self._generate_valuation_analysis(data)
            report += self._generate_risk_assessment(data)
            report += self._generate_investment_thesis(data)
            report += self._generate_report_footer()

            return report

        except Exception as e:
            return f"Research report generation failed for {symbol}: {str(e)}"

    def _generate_report_header(self, data: Dict) -> str:
        """Generate report header with company info"""
        return f"""
# Investment Research Report
## {data.get('Name', 'Unknown Company')} ({data.get('Symbol', 'N/A')})

**Report Date:** {datetime.now().strftime('%Y-%m-%d')}
**Sector:** {data.get('Sector', 'N/A')}
**Industry:** {data.get('Industry', 'N/A')}
**Exchange:** {data.get('Exchange', 'N/A')}

---

"""

    def _generate_company_overview(self, data: Dict) -> str:
        """Generate company overview section"""
        description = data.get('Description', 'No description available.')

        return f"""
## Company Overview

{description}

**Key Details:**
- Country: {data.get('Country', 'N/A')}
- Currency: {data.get('Currency', 'N/A')}
- Fiscal Year End: {data.get('FiscalYearEnd', 'N/A')}
- Address: {data.get('Address', 'N/A')}

---

"""

    def _generate_financial_analysis(self, data: Dict) -> str:
        """Generate financial metrics analysis"""
        def format_large_number(value_str):
            try:
                value = float(value_str)
                if value >= 1e12:
                    return f"${value/1e12:.2f}T"
                elif value >= 1e9:
                    return f"${value/1e9:.2f}B"
                elif value >= 1e6:
                    return f"${value/1e6:.2f}M"
                else:
                    return f"${value:,.0f}"
            except:
                return value_str

        def format_percentage(value_str):
            try:
                value = float(value_str)
                return f"{value*100:.2f}%"
            except:
                return value_str

        return f"""
## Financial Analysis

### Key Metrics
- **Market Capitalization:** {format_large_number(data.get('MarketCapitalization', '0'))}
- **Revenue (TTM):** {format_large_number(data.get('RevenueTTM', '0'))}
- **EBITDA:** {format_large_number(data.get('EBITDA', '0'))}
- **Earnings Per Share:** ${data.get('EPS', 'N/A')}
- **Book Value Per Share:** ${data.get('BookValue', 'N/A')}

### Profitability Metrics
- **Profit Margin:** {format_percentage(data.get('ProfitMargin', '0'))}
- **Operating Margin (TTM):** {format_percentage(data.get('OperatingMarginTTM', '0'))}
- **Return on Assets (TTM):** {format_percentage(data.get('ReturnOnAssetsTTM', '0'))}
- **Return on Equity (TTM):** {format_percentage(data.get('ReturnOnEquityTTM', '0'))}

### Dividend Information
- **Dividend Per Share:** ${data.get('DividendPerShare', '0')}
- **Dividend Yield:** {format_percentage(data.get('DividendYield', '0'))}
- **Dividend Date:** {data.get('DividendDate', 'N/A')}
- **Ex-Dividend Date:** {data.get('ExDividendDate', 'N/A')}

---

"""

    def _generate_valuation_analysis(self, data: Dict) -> str:
        """Generate valuation metrics analysis"""
        def safe_float(value, default="N/A"):
            try:
                return float(value)
            except:
                return default

        pe_ratio = safe_float(data.get('PERatio'))
        pb_ratio = safe_float(data.get('PriceToBookRatio'))
        ps_ratio = safe_float(data.get('PriceToSalesRatioTTM'))

        # Valuation assessment
        valuation_notes = []

        if isinstance(pe_ratio, float):
            if pe_ratio < 15:
                valuation_notes.append("P/E ratio suggests potential undervaluation")
            elif pe_ratio > 30:
                valuation_notes.append("P/E ratio indicates high valuation premium")

        if isinstance(pb_ratio, float):
            if pb_ratio < 1:
                valuation_notes.append("Trading below book value - potential value opportunity")
            elif pb_ratio > 3:
                valuation_notes.append("High price-to-book ratio suggests growth premium")

        valuation_assessment = "\n".join([f"- {note}" for note in valuation_notes]) if valuation_notes else "- No specific valuation concerns identified"

        return f"""
## Valuation Analysis

### Valuation Ratios
- **P/E Ratio:** {pe_ratio if isinstance(pe_ratio, float) else 'N/A'}
- **PEG Ratio:** {data.get('PEGRatio', 'N/A')}
- **Price-to-Book:** {pb_ratio if isinstance(pb_ratio, float) else 'N/A'}
- **Price-to-Sales (TTM):** {ps_ratio if isinstance(ps_ratio, float) else 'N/A'}
- **EV/Revenue:** {data.get('EVToRevenue', 'N/A')}
- **EV/EBITDA:** {data.get('EVToEBITDA', 'N/A')}

### Valuation Assessment
{valuation_assessment}

### Analyst Expectations
- **Target Price:** ${data.get('AnalystTargetPrice', 'N/A')}
- **Forward P/E:** {data.get('ForwardPE', 'N/A')}

---

"""

    def _generate_risk_assessment(self, data: Dict) -> str:
        """Generate risk assessment section"""
        beta = data.get('Beta', 'N/A')

        risk_factors = []

        # Beta risk
        try:
            beta_val = float(beta)
            if beta_val > 1.5:
                risk_factors.append("High volatility risk - Beta above 1.5 indicates significant price sensitivity to market movements")
            elif beta_val < 0.5:
                risk_factors.append("Low market correlation - Beta below 0.5 may indicate defensive characteristics")
        except:
            pass

        # Sector-specific risks
        sector = data.get('Sector', '').lower()
        if 'technology' in sector:
            risk_factors.append("Technology sector exposure to rapid innovation cycles and regulatory changes")
        elif 'financial' in sector:
            risk_factors.append("Financial sector exposure to interest rate changes and regulatory oversight")
        elif 'energy' in sector:
            risk_factors.append("Energy sector exposure to commodity price volatility and environmental regulations")

        # Market cap risk
        try:
            market_cap = float(data.get('MarketCapitalization', 0))
            if market_cap < 1e9:  # Less than $1B
                risk_factors.append("Small-cap risk - Lower liquidity and higher volatility potential")
        except:
            pass

        risk_summary = "\n".join([f"- {risk}" for risk in risk_factors]) if risk_factors else "- No specific risk factors identified at this time"

        return f"""
## Risk Assessment

### Market Risk Indicators
- **Beta:** {beta}
- **52-Week Range:** ${data.get('52WeekLow', 'N/A')} - ${data.get('52WeekHigh', 'N/A')}

### Identified Risk Factors
{risk_summary}

---

"""

    def _generate_investment_thesis(self, data: Dict) -> str:
        """Generate investment thesis based on data"""
        thesis_points = []

        # Profitability thesis
        try:
            profit_margin = float(data.get('ProfitMargin', 0))
            if profit_margin > 0.15:
                thesis_points.append(f"Strong profitability with {profit_margin*100:.1f}% profit margins")
        except:
            pass

        # Growth thesis
        try:
            earnings_growth = float(data.get('QuarterlyEarningsGrowthYOY', 0))
            if earnings_growth > 0.1:
                thesis_points.append(f"Positive earnings momentum with {earnings_growth*100:.1f}% YoY growth")
        except:
            pass

        # Dividend thesis
        try:
            div_yield = float(data.get('DividendYield', 0))
            if div_yield > 0.02:
                thesis_points.append(f"Income generation through {div_yield*100:.2f}% dividend yield")
        except:
            pass

        # Market position thesis
        try:
            market_cap = float(data.get('MarketCapitalization', 0))
            if market_cap > 1e11:  # $100B+
                thesis_points.append("Large-cap stability and market leadership position")
        except:
            pass

        investment_thesis = "\n".join([f"- {point}" for point in thesis_points]) if thesis_points else "- Investment thesis requires further fundamental analysis"

        return f"""
## Investment Thesis

### Key Investment Highlights
{investment_thesis}

### Recommendation Framework
This analysis provides a foundation for investment decisions. Consider the following:

1. **Risk Tolerance:** Assess whether the identified risks align with your investment objectives
2. **Time Horizon:** Consider how the company's fundamentals match your investment timeline
3. **Portfolio Diversification:** Evaluate how this position fits within your overall portfolio strategy
4. **Market Conditions:** Consider current market environment and sector rotation trends

---

"""

    def _generate_report_footer(self) -> str:
        """Generate report footer with disclaimers"""
        return """
## Important Disclaimers

**Data Source:** This report is based on data from Alpha Vantage API and should be verified with additional sources.

**Not Investment Advice:** This report is for informational purposes only and does not constitute investment advice. Please consult with a qualified financial advisor before making investment decisions.

**Risk Warning:** All investments carry risk, including the potential loss of principal. Past performance does not guarantee future results.

**Report Limitations:** This automated analysis may not capture all relevant factors affecting the investment. Additional due diligence is recommended.

---
*Report generated by Simple MCP Market Data Server*
"""

# Usage example for generating research reports
async def generate_research_reports():
    client = SimpleMCPClient("./simple-mcp")
    research_bot = InvestmentResearchBot(client)

    try:
        await client.start()

        # Generate reports for multiple stocks
        symbols = ["AAPL", "MSFT", "GOOGL"]

        for symbol in symbols:
            try:
                print(f"Generating research report for {symbol}...")
                report = await research_bot.generate_research_report(symbol)

                # Save to file
                filename = f"{symbol}_research_report_{datetime.now().strftime('%Y%m%d')}.md"
                with open(filename, 'w') as f:
                    f.write(report)

                print(f"Report saved to {filename}")

            except Exception as e:
                print(f"Failed to generate report for {symbol}: {e}")

    finally:
        await client.close()

if __name__ == "__main__":
    asyncio.run(generate_research_reports())
```

## Portfolio Screening Tool

### Quantitative Stock Screener

```python
import asyncio
from typing import List, Dict, Callable
from dataclasses import dataclass

@dataclass
class ScreeningCriteria:
    name: str
    description: str
    filter_function: Callable[[Dict], bool]

@dataclass
class ScreeningResult:
    symbol: str
    name: str
    matches: List[str]
    key_metrics: Dict[str, float]
    score: float

class PortfolioScreener:
    """Quantitative stock screening tool"""

    def __init__(self, mcp_client):
        self.client = mcp_client
        self.screening_criteria = self._initialize_criteria()

    def _initialize_criteria(self) -> List[ScreeningCriteria]:
        """Initialize common screening criteria"""
        return [
            ScreeningCriteria(
                "Low P/E Value",
                "P/E ratio between 5 and 15",
                lambda data: 5 <= self._safe_float(data.get('PERatio')) <= 15
            ),
            ScreeningCriteria(
                "High Dividend Yield",
                "Dividend yield above 3%",
                lambda data: self._safe_float(data.get('DividendYield')) > 0.03
            ),
            ScreeningCriteria(
                "Strong Profitability",
                "Profit margin above 15%",
                lambda data: self._safe_float(data.get('ProfitMargin')) > 0.15
            ),
            ScreeningCriteria(
                "High ROE",
                "Return on Equity above 15%",
                lambda data: self._safe_float(data.get('ReturnOnEquityTTM')) > 0.15
            ),
            ScreeningCriteria(
                "Large Cap",
                "Market cap above $10B",
                lambda data: self._safe_float(data.get('MarketCapitalization')) > 1e10
            ),
            ScreeningCriteria(
                "Positive Earnings Growth",
                "Quarterly earnings growth above 10%",
                lambda data: self._safe_float(data.get('QuarterlyEarningsGrowthYOY')) > 0.10
            ),
            ScreeningCriteria(
                "Low Beta",
                "Beta less than 1.2 (lower volatility)",
                lambda data: self._safe_float(data.get('Beta'), 1.0) < 1.2
            ),
            ScreeningCriteria(
                "Strong Balance Sheet",
                "Book value per share above $5",
                lambda data: self._safe_float(data.get('BookValue')) > 5.0
            )
        ]

    def _safe_float(self, value, default=0.0):
        """Safely convert to float"""
        try:
            return float(value) if value else default
        except (ValueError, TypeError):
            return default

    async def screen_stocks(self, symbols: List[str],
                           selected_criteria: List[str] = None,
                           min_matches: int = 2) -> List[ScreeningResult]:
        """Screen stocks against criteria"""
        results = []

        # Use all criteria if none specified
        if selected_criteria is None:
            criteria_to_use = self.screening_criteria
        else:
            criteria_to_use = [c for c in self.screening_criteria if c.name in selected_criteria]

        for symbol in symbols:
            try:
                data = await self.client.get_stock_data(symbol)

                # Check which criteria are met
                matches = []
                for criterion in criteria_to_use:
                    if criterion.filter_function(data):
                        matches.append(criterion.name)

                # Only include if minimum matches are met
                if len(matches) >= min_matches:
                    score = len(matches) / len(criteria_to_use)

                    result = ScreeningResult(
                        symbol=symbol,
                        name=data.get('Name', 'Unknown'),
                        matches=matches,
                        key_metrics=self._extract_key_metrics(data),
                        score=score
                    )
                    results.append(result)

            except Exception as e:
                print(f"Error screening {symbol}: {e}")
                continue

        # Sort by score descending
        return sorted(results, key=lambda x: x.score, reverse=True)

    def _extract_key_metrics(self, data: Dict) -> Dict[str, float]:
        """Extract key metrics for display"""
        return {
            "P/E Ratio": self._safe_float(data.get('PERatio')),
            "Dividend Yield": self._safe_float(data.get('DividendYield')),
            "Profit Margin": self._safe_float(data.get('ProfitMargin')),
            "ROE": self._safe_float(data.get('ReturnOnEquityTTM')),
            "Market Cap (B)": self._safe_float(data.get('MarketCapitalization')) / 1e9,
            "Beta": self._safe_float(data.get('Beta'))
        }

    def print_screening_results(self, results: List[ScreeningResult]):
        """Print formatted screening results"""
        if not results:
            print("No stocks passed the screening criteria.")
            return

        print(f"\n{'='*60}")
        print(f"SCREENING RESULTS - {len(results)} stocks found")
        print(f"{'='*60}")

        for i, result in enumerate(results, 1):
            print(f"\n{i}. {result.name} ({result.symbol})")
            print(f"   Score: {result.score:.1%} ({len(result.matches)}/{len(self.screening_criteria)} criteria met)")
            print(f"   Criteria matched: {', '.join(result.matches)}")

            print("   Key Metrics:")
            for metric, value in result.key_metrics.items():
                if metric == "Market Cap (B)":
                    print(f"     {metric}: ${value:.1f}B")
                elif metric in ["Dividend Yield", "Profit Margin", "ROE"]:
                    print(f"     {metric}: {value*100:.1f}%")
                else:
                    print(f"     {metric}: {
