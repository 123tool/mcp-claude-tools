package mcp

// JSONRPCRequest merepresentasikan request masuk dari klien MCP
type JSONRPCRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	Method  string         `json:"method"`
	Params  ParamsPayload  `json:"params"`
	ID      interface{}    `json:"id"`
}

type ParamsPayload struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// JSONRPCResponse merepresentasikan respons standar MCP
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ToolResult struct {
	Content []TextContent `json:"content"`
	IsError bool          `json:"isError"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
