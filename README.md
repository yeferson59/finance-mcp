simple-mcp/README.md

# Simple MCP Market Data Server

**Simple MCP Market Data** is an MCP (Model Context Protocol) server implemented in Go that provides tools for querying real-time stock market financial data via the Alpha Vantage API.

## 🏗️ Project Architecture

This project implements an MCP server that exposes financial data query functionalities as tools usable by language models and other MCP clients.

### Main Components

- **MCP Server**: Protocol implementation using `github.com/modelcontextprotocol/go-sdk`
- **Alpha Vantage Client**: Integration with the Alpha Vantage API to fetch financial data
- **Data Models**: Go structures with JSON Schema annotations for validation
- **Tools**: MCP tool implementations specific to financial data

## 🔧 Features

### Available MCP Tools

#### `get-stock`

- **Purpose**: Retrieves detailed information about a specific stock
- **Parameters**:
  - `symbol` (string): Stock symbol (e.g., "AAPL", "GOOGL", "MSFT")
- **Response**: JSON object with comprehensive company and stock data
- **API Used**: Alpha Vantage OVERVIEW function

### Returned Data

The `get-stock` tool provides comprehensive information including:

- Basic company information (name, description, sector, industry)
- Financial metrics (market cap, P/E ratio, dividends)
- Market data (price, volume, 52-week range)
- Fundamental data (EPS, beta, ROE, margins)

## 📁 Project Structure

```
simple-mcp/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── config/              # Configuration and environment variables
│   ├── models/              # Data models and structures
│   │   ├── inputs.go        # Input structures with JSON Schema
│   │   ├── outputs.go       # Response structures
│   │   └── stock.go         # Stock-specific models
│   ├── prompts/             # Prompts and templates (reserved)
│   └── tools/               # MCP tool implementations
│       └── overview_stock.go # Stock query tool
├── pkg/
│   └── env.go              # Utilities for environment variable management
├── go.mod                  # Project dependencies
├── go.sum                  # Dependency checksums
└── main                    # Compiled binary
```

### Component Descriptions

#### `cmd/main.go`

Initializes the MCP server, registers available tools, and manages stdio transport for communication.

#### `internal/tools/overview_stock.go`

Implements the `get-stock` tool that:

1. Accepts a stock symbol as input
2. Makes an HTTP request to the Alpha Vantage API
3. Deserializes the JSON response
4. Returns structured data

#### `internal/models/`

Defines data structures with JSON Schema annotations:

- `SymbolInput`: Tool input parameters
- `OverviewOutput`: Output data structure
- Automatic type and parameter validation

## 🚀 Installation & Setup

### Prerequisites

- Go 1.25.1 or higher
- Alpha Vantage API Key (free at https://www.alphavantage.co/support/#api-key)

### Installation Steps

1. **Clone the repository:**

```bash
git clone <repository-url>
cd simple-mcp
```

2. **Install dependencies:**

```bash
go mod tidy
```

3. **Configure environment variables:**
   Create a `.env` file in the root:

```env
API_URL=https://www.alphavantage.co
API_KEY=your_alpha_vantage_api_key_here
```

4. **Build (optional):**

```bash
go build -o simple-mcp cmd/main.go
```

## 🛠️ Usage

### Run the Server

```bash
go run cmd/main.go
```

The server runs using stdio transport, allowing direct communication with MCP clients.

### Example Usage with MCP Client

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "get-stock",
    "arguments": {
      "symbol": "AAPL"
    }
  }
}
```

### Example Response

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\n  \"symbol\": \"AAPL\",\n  \"name\": \"Apple Inc\",\n  \"description\": \"Apple Inc. designs, manufactures, and markets smartphones...\",\n  \"sector\": \"Technology\",\n  \"industry\": \"Consumer Electronics\",\n  \"market_cap\": \"3000000000000\",\n  ...\n}"
      }
    ]
  }
}
```

## 🔌 Integration with AI Models

This MCP server is designed to be used by:

- **Large Language Models (LLMs)** needing real-time financial data access
- **Automated financial analysis applications**
- **Finance and investment chatbots**
- **Investment recommendation systems**

### Typical Use Cases

1. **Fundamental Analysis**: Retrieve key metrics for stock analysis
2. **Market Research**: Query sector and industry information
3. **Stock Screening**: Filter stocks by specific criteria
4. **Financial Education**: Provide real data for learning

## 🔧 Main Dependencies

- `github.com/modelcontextprotocol/go-sdk v0.6.0`: Official MCP SDK for Go
- `github.com/joho/godotenv v1.5.1`: Environment variable management
- `github.com/google/jsonschema-go v0.2.3`: JSON schema validation

## 📊 External API

### Alpha Vantage Integration

- **Endpoint**: `https://www.alphavantage.co/query`
- **Function used**: `OVERVIEW`
- **Response format**: JSON
- **Limits**: 25 requests/day (free plan), 500 requests/day (premium plan)

## 🔒 Security Considerations

- API Key stored in environment variables (not hardcoded)
- Input validation using JSON Schema
- HTTP and network error handling
- Implicit timeouts in HTTP requests

## 🤝 Contributions

To contribute to the project:

1. Fork the repository
2. Create a branch for your feature (`git checkout -b feature/new-feature`)
3. Commit your changes (`git commit -am 'Add new feature'`)
4. Push to your branch (`git push origin feature/new-feature`)
5. Create a Pull Request

### Future Roadmap

- [ ] Add tool for historical price data
- [ ] Implement tool for stock search
- [ ] Add cryptocurrency support
- [ ] Implement caching to reduce API calls
- [ ] Add metrics and logging

## 📝 License & Credits

**Developed by**: Yeferson Toloza Contreras

**APIs used**:

- [Alpha Vantage](https://www.alphavantage.co/) - Real-time financial data
- [Model Context Protocol](https://github.com/modelcontextprotocol/go-sdk) - Communication framework

**License**: [Specify license]

---

For technical support or integration questions, consult the official MCP documentation or open an issue in this repository.
