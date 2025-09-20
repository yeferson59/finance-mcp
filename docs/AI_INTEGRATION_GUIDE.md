# AI Integration Guide for Simple MCP Market Data Server

This document provides comprehensive technical information for AI models, automated systems, and developers integrating with the Simple MCP Market Data Server.

## Table of Contents

1. [Protocol Overview](#protocol-overview)
2. [Tool Specifications](#tool-specifications)
3. [JSON Schemas](#json-schemas)
4. [Request/Response Examples](#requestresponse-examples)
5. [Error Handling](#error-handling)
6. [AI Integration Patterns](#ai-integration-patterns)
7. [Rate Limits and Best Practices](#rate-limits-and-best-practices)
8. [Troubleshooting](#troubleshooting)

## Protocol Overview

### MCP (Model Context Protocol) Basics

This server implements MCP version 2024-11-05 using JSON-RPC 2.0 over stdio transport.

- **Transport**: stdio (stdin/stdout)
- **Protocol**: JSON-RPC 2.0
- **Data Format**: JSON
- **Communication**: Bidirectional, client-initiated

### Server Capabilities

```json
{
  "capabilities": {
    "tools": {
      "listChanged": true
    }
  },
  "protocolVersion": "2024-11-05",
  "serverInfo": {
    "name": "simple-mcp-market-data",
    "version": "1.0.0"
  }
}
```

## Tool Specifications

### Tool: `get-stock`

**Purpose**: Retrieve comprehensive stock market data and company information.

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "symbol": {
      "type": "string",
      "description": "Stock ticker symbol (e.g. AAPL for Apple Inc, GOOGL for Alphabet/Google, MSFT for Microsoft). Must be a valid stock symbol recognized by major exchanges.",
      "pattern": "^[A-Za-z]{1,10}$",
      "examples": ["AAPL", "GOOGL", "MSFT", "AMZN", "TSLA", "NVDA", "META", "BRK.B"]
    }
  },
  "required": ["symbol"],
  "additionalProperties": false
}
```

**Output Data Types**:
- All monetary values: String (to preserve precision)
- All ratios and percentages: String
- Dates: String in YYYY-MM-DD format
- Text fields: String (may contain null/empty values)

## JSON Schemas

### Complete Request Schema

```json
{
  "jsonrpc": "2.0",
  "id": "unique-request-id",
  "method": "tools/call",
  "params": {
    "name": "get-stock",
    "arguments": {
      "symbol": "AAPL"
    }
  }
}
```

### Complete Response Schema

```json
{
  "jsonrpc": "2.0",
  "id": "unique-request-id",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"Symbol\":\"AAPL\",\"Name\":\"Apple Inc\",\"Description\":\"Apple Inc. designs, manufactures, and markets smartphones, personal computers, tablets, wearables, and accessories worldwide...\",\"Sector\":\"Technology\",\"Industry\":\"Consumer Electronics\",\"MarketCapitalization\":\"3000000000000\",\"PERatio\":\"28.5\",\"EPS\":\"6.05\",\"DividendYield\":\"0.45\"}"
      }
    ],
    "isError": false
  }
}
```

### Error Response Schema

```json
{
  "jsonrpc": "2.0",
  "id": "unique-request-id",
  "error": {
    "code": -1,
    "message": "Tool execution failed",
    "data": {
      "details": "specific error description"
    }
  }
}
```

## Request/Response Examples

### Example 1: Apple Inc. (AAPL)

**Request**:
```json
{
  "jsonrpc": "2.0",
  "id": "req-001",
  "method": "tools/call",
  "params": {
    "name": "get-stock",
    "arguments": {
      "symbol": "AAPL"
    }
  }
}
```

**Response** (abbreviated):
```json
{
  "jsonrpc": "2.0",
  "id": "req-001",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"Symbol\":\"AAPL\",\"Name\":\"Apple Inc\",\"Description\":\"Apple Inc. designs, manufactures, and markets smartphones, personal computers, tablets, wearables, and accessories worldwide.\",\"Country\":\"USA\",\"Sector\":\"Technology\",\"Industry\":\"Consumer Electronics\",\"MarketCapitalization\":\"3000000000000\",\"EBITDA\":\"125000000000\",\"PERatio\":\"28.50\",\"PEGRatio\":\"2.15\",\"BookValue\":\"4.25\",\"DividendPerShare\":\"0.96\",\"DividendYield\":\"0.0045\",\"EPS\":\"6.05\",\"RevenuePerShareTTM\":\"24.32\",\"ProfitMargin\":\"0.258\",\"OperatingMarginTTM\":\"0.298\",\"ReturnOnAssetsTTM\":\"0.204\",\"ReturnOnEquityTTM\":\"1.479\",\"RevenueTTM\":\"394000000000\",\"GrossProfitTTM\":\"171000000000\",\"DilutedEPSTTM\":\"6.05\",\"QuarterlyEarningsGrowthYOY\":\"0.072\",\"QuarterlyRevenueGrowthYOY\":\"0.045\",\"AnalystTargetPrice\":\"185.50\",\"52WeekHigh\":\"199.62\",\"52WeekLow\":\"164.08\",\"50DayMovingAverage\":\"178.45\",\"200DayMovingAverage\":\"175.32\",\"SharesOutstanding\":\"15550000000\",\"DividendDate\":\"2024-02-16\",\"ExDividendDate\":\"2024-02-09\"}"
      }
    ],
    "isError": false
  }
}
```

### Example 2: Error Case - Invalid Symbol

**Request**:
```json
{
  "jsonrpc": "2.0",
  "id": "req-002",
  "method": "tools/call",
  "params": {
    "name": "get-stock",
    "arguments": {
      "symbol": "INVALID"
    }
  }
}
```

**Response**:
```json
{
  "jsonrpc": "2.0",
  "id": "req-002",
  "error": {
    "code": -1,
    "message": "Tool execution failed",
    "data": {
      "details": "no stock data found for symbol 'INVALID': symbol may not exist or may not be supported by Alpha Vantage"
    }
  }
}
```

## Error Handling

### Common Error Categories

1. **Input Validation Errors**:
   - Empty symbol: `"stock symbol cannot be empty"`
   - Invalid format: `"stock symbol 'TOOLONGSYMBOL' appears to be invalid (too long)"`

2. **API Errors**:
   - Rate limit: `"Alpha Vantage API note for symbol 'AAPL': Thank you for using Alpha Vantage! Our standard API call frequency is 25 requests per day"`
   - Invalid symbol: `"no stock data found for symbol 'XYZ'"`
   - Network issues: `"failed to fetch stock data for symbol 'AAPL'"`

3. **System Errors**:
   - Parsing errors: `"failed to parse stock data response"`
   - HTTP errors: `"Alpha Vantage API returned status 500"`

### Error Recovery Strategies

For AI systems, implement these strategies:

1. **Rate Limiting**: Wait 24 hours before retrying if rate limit exceeded
2. **Invalid Symbols**: Try alternative ticker formats or exchange-specific symbols
3. **Network Issues**: Implement exponential backoff with 3 retry attempts
4. **Parsing Errors**: Log the raw response and report the issue

## AI Integration Patterns

### Pattern 1: Financial Analysis Assistant

```python
# Pseudo-code for AI integration
async def analyze_stock(symbol: str) -> str:
    try:
        response = await mcp_client.call_tool("get-stock", {"symbol": symbol})
        data = json.loads(response.content[0].text)

        # Extract key metrics for analysis
        metrics = {
            "pe_ratio": float(data.get("PERatio", 0)),
            "dividend_yield": float(data.get("DividendYield", 0)),
            "profit_margin": float(data.get("ProfitMargin", 0)),
            "market_cap": int(data.get("MarketCapitalization", 0))
        }

        # Generate analysis
        return generate_financial_analysis(data, metrics)

    except Exception as e:
        return f"Unable to analyze {symbol}: {str(e)}"
```

### Pattern 2: Portfolio Screening

```python
async def screen_stocks(symbols: List[str], criteria: Dict) -> List[str]:
    results = []

    for symbol in symbols:
        try:
            response = await mcp_client.call_tool("get-stock", {"symbol": symbol})
            data = json.loads(response.content[0].text)

            # Apply screening criteria
            if meets_criteria(data, criteria):
                results.append(symbol)

        except Exception:
            continue  # Skip invalid symbols

    return results
```

### Pattern 3: Real-time Data Integration

```python
class StockDataCache:
    def __init__(self):
        self.cache = {}
        self.last_updated = {}

    async def get_stock_data(self, symbol: str, max_age_hours: int = 24):
        # Check cache freshness
        if self.is_cache_valid(symbol, max_age_hours):
            return self.cache[symbol]

        # Fetch fresh data
        response = await mcp_client.call_tool("get-stock", {"symbol": symbol})
        data = json.loads(response.content[0].text)

        # Update cache
        self.cache[symbol] = data
        self.last_updated[symbol] = datetime.now()

        return data
```

## Rate Limits and Best Practices

### Alpha Vantage API Limits

- **Free Tier**: 25 requests per day, 5 requests per minute
- **Premium Tier**: 75-1200 requests per minute (depending on plan)

### Best Practices for AI Systems

1. **Caching Strategy**:
   ```python
   # Cache stock data for at least 1 hour for active trading hours
   # Cache for 24 hours for after-hours or weekend analysis
   CACHE_DURATION_ACTIVE = 3600    # 1 hour
   CACHE_DURATION_INACTIVE = 86400  # 24 hours
   ```

2. **Batch Processing**:
   ```python
   # Process stocks in batches with delays
   for batch in chunked(symbols, batch_size=20):
       for symbol in batch:
           await process_stock(symbol)
       await asyncio.sleep(60)  # Wait 1 minute between batches
   ```

3. **Error Resilience**:
   ```python
   async def robust_stock_fetch(symbol: str, max_retries: int = 3):
       for attempt in range(max_retries):
           try:
               return await fetch_stock_data(symbol)
           except RateLimitError:
               await asyncio.sleep(3600)  # Wait 1 hour
           except NetworkError:
               await asyncio.sleep(2 ** attempt)  # Exponential backoff
       raise Exception(f"Failed to fetch {symbol} after {max_retries} attempts")
   ```

## Troubleshooting

### Common Issues and Solutions

1. **"No stock data found"**:
   - Verify symbol spelling and format
   - Try alternative formats (e.g., "BRK.B" vs "BRKB")
   - Check if symbol exists on major exchanges

2. **Rate limit exceeded**:
   - Implement proper caching
   - Reduce request frequency
   - Consider upgrading API plan

3. **Empty response fields**:
   - Some stocks may not have all data fields
   - Always check for null/empty values before processing
   - Use fallback values or skip incomplete data

4. **JSON parsing errors**:
   - Alpha Vantage occasionally returns HTML error pages
   - Implement response format validation
   - Log raw responses for debugging

### Debugging Tools

```python
def debug_stock_response(symbol: str, response: str):
    """Debug helper for analyzing API responses"""
    print(f"=== DEBUG: {symbol} ===")
    print(f"Response length: {len(response)}")
    print(f"First 200 chars: {response[:200]}")

    try:
        data = json.loads(response)
        print(f"JSON fields: {list(data.keys())}")
        print(f"Symbol field: {data.get('Symbol', 'MISSING')}")
        print(f"Name field: {data.get('Name', 'MISSING')}")
        print(f"Error fields: {data.get('Error Message', 'NONE')}")
    except json.JSONDecodeError as e:
        print(f"JSON parsing failed: {e}")
```

### Performance Monitoring

```python
import time
from typing import Dict

class MCPPerformanceMonitor:
    def __init__(self):
        self.request_times = []
        self.error_counts = {}

    def record_request(self, symbol: str, duration: float, success: bool):
        self.request_times.append(duration)

        if not success:
            self.error_counts[symbol] = self.error_counts.get(symbol, 0) + 1

    def get_stats(self) -> Dict:
        if not self.request_times:
            return {"avg_response_time": 0, "total_errors": 0}

        return {
            "avg_response_time": sum(self.request_times) / len(self.request_times),
            "max_response_time": max(self.request_times),
            "min_response_time": min(self.request_times),
            "total_errors": sum(self.error_counts.values()),
            "error_breakdown": self.error_counts
        }
```

## Integration Examples for Specific AI Use Cases

### Investment Research Assistant

```python
async def research_stock(symbol: str) -> str:
    """Generate comprehensive investment research report"""
    data = await get_stock_data(symbol)

    analysis_points = []

    # Valuation analysis
    pe_ratio = float(data.get("PERatio", 0))
    if pe_ratio > 0:
        if pe_ratio < 15:
            analysis_points.append("Stock appears undervalued based on P/E ratio")
        elif pe_ratio > 30:
            analysis_points.append("Stock may be overvalued based on P/E ratio")

    # Dividend analysis
    div_yield = float(data.get("DividendYield", 0))
    if div_yield > 0.03:  # 3%
        analysis_points.append(f"Attractive dividend yield of {div_yield*100:.2f}%")

    # Growth analysis
    earnings_growth = float(data.get("QuarterlyEarningsGrowthYOY", 0))
    if earnings_growth > 0.1:  # 10%
        analysis_points.append("Strong earnings growth momentum")

    return f"Analysis for {data.get('Name', symbol)}:\n" + "\n".join(analysis_points)
```

### Risk Assessment Tool

```python
def assess_risk(stock_data: Dict) -> Dict[str, str]:
    """Assess various risk factors for a stock"""
    risks = {}

    # Volatility risk
    beta = float(stock_data.get("Beta", 1.0))
    if beta > 1.5:
        risks["volatility"] = "High - Stock is significantly more volatile than market"
    elif beta < 0.5:
        risks["volatility"] = "Low - Stock is less volatile than market"

    # Valuation risk
    pe_ratio = float(stock_data.get("PERatio", 0))
    if pe_ratio > 50:
        risks["valuation"] = "High - Stock trading at premium valuation"

    # Liquidity risk (based on market cap)
    market_cap = int(stock_data.get("MarketCapitalization", 0))
    if market_cap < 1000000000:  # Less than $1B
        risks["liquidity"] = "Medium - Small cap stock may have liquidity constraints"

    return risks
```

This guide provides comprehensive information for AI systems to effectively integrate with and utilize the Simple MCP Market Data Server. For additional support or advanced integration scenarios, refer to the main project documentation or submit issues to the project repository.
