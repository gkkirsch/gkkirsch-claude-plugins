# Agent Patterns Reference

Reference for AI agent architectures — ReAct, plan-and-execute, reflection, tool calling, function calling, and MCP patterns.

---

## Agent Architecture Patterns

### Pattern Comparison

| Pattern | Best For | Complexity | Reliability | Latency |
|---------|----------|------------|-------------|---------|
| ReAct | General tool use, Q&A | Low | High | Medium |
| Plan-and-Execute | Complex multi-step tasks | Medium | Medium | High |
| Reflection | Quality-critical outputs | Medium | High | High |
| Tree-of-Thought | Complex reasoning | High | High | Very High |
| Self-Ask | Multi-hop questions | Low | Medium | Medium |
| LATS (Language Agent Tree Search) | Complex planning | High | Very High | Very High |

### When to Use Each Pattern

```
ReAct:
  - User asks a question → agent searches → answers
  - Simple tool-use workflows (1-3 tools per query)
  - When you need predictable, traceable behavior
  - 80% of use cases

Plan-and-Execute:
  - Multi-step research tasks
  - Tasks requiring 5+ distinct actions
  - When order of operations matters
  - When you want human review of the plan before execution

Reflection:
  - Code generation (generate → review → fix)
  - Content creation (draft → critique → revise)
  - When output quality matters more than speed
  - When you have clear evaluation criteria

Tree-of-Thought:
  - Mathematical reasoning
  - Strategic planning
  - When the problem has multiple valid approaches
  - When wrong early choices cascade into bad results

Self-Ask:
  - Multi-hop factual questions
  - Questions that decompose into sub-questions
  - "Who is the CEO of the company that makes the iPhone?"
```

---

## ReAct Pattern

### Architecture

```
┌──────────────────────────────────────────────────┐
│                  ReAct Loop                       │
│                                                    │
│  User Query ─→ [THOUGHT] ─→ [ACTION] ─→          │
│                    ↑                     ↓         │
│                    └── [THOUGHT] ←── [OBSERVATION] │
│                    ↑                     ↓         │
│                    └── ... until done              │
│                                                    │
│  Final Answer ←── [THOUGHT: I have enough info]   │
└──────────────────────────────────────────────────┘
```

### ReAct Prompt Template

```
You are a helpful assistant with access to the following tools:

{tool_descriptions}

When answering, follow this cycle:
1. THOUGHT: Reason about what you need to know or do
2. ACTION: Call a tool if needed
3. OBSERVATION: Review the tool result
4. Repeat until you can answer confidently

Rules:
- Think before acting
- Use the minimum number of tool calls necessary
- If a tool returns an error, try a different approach
- When you have enough information, provide your final answer
```

### Common ReAct Failure Modes

```
Problem: Agent loops (keeps calling same tool)
Fix: Set max_steps limit, add "if you've tried X twice, try a different approach"

Problem: Agent doesn't use tools when it should
Fix: Add explicit instruction "always verify facts using tools, don't rely on memory"

Problem: Agent uses wrong tool
Fix: Improve tool descriptions, add examples of when to use each tool

Problem: Agent provides answer too early (before getting enough info)
Fix: Add "gather at least N sources before answering" or quality check step

Problem: Agent hallucinates tool names
Fix: Use function calling (structured tool use) instead of text-based tool calls
```

---

## Plan-and-Execute Pattern

### Architecture

```
┌───────────────────────────────────────────────────────────┐
│                  Plan-and-Execute                           │
│                                                             │
│  Query ─→ [PLANNER] ─→ Plan: [Step 1, Step 2, ...]        │
│                              │                              │
│              ┌───────────────┘                              │
│              ↓                                              │
│  [EXECUTOR: Step 1] ─→ Result 1                            │
│              ↓                                              │
│  [RE-PLAN?] ─→ Yes → [PLANNER: revise remaining steps]    │
│       │        No                                           │
│       ↓                                                     │
│  [EXECUTOR: Step 2] ─→ Result 2                            │
│       ...                                                   │
│              ↓                                              │
│  [SYNTHESIZER] ─→ Final Answer                             │
└───────────────────────────────────────────────────────────┘
```

### Plan-and-Execute Best Practices

```
Planning:
  - Plans should have 3-7 steps (too few = missed steps, too many = over-planning)
  - Each step should be independently executable
  - Include expected output for each step
  - Add "verify" steps after critical operations

Execution:
  - Execute one step at a time
  - Pass results from previous steps as context
  - Check for failures after each step
  - Re-plan if results deviate from expectations

Re-planning triggers:
  - A step fails
  - Results contradict assumptions
  - New information suggests a better path
  - Mid-execution user input changes requirements
```

---

## Reflection Pattern

### Architecture

```
┌──────────────────────────────────────────────┐
│              Reflection Loop                  │
│                                                │
│  Query ─→ [GENERATE] ─→ Output v1            │
│                              ↓                 │
│            [CRITIQUE] ─→ Feedback              │
│                ↓                               │
│           Score >= threshold? ─→ Yes → Done    │
│                ↓ No                            │
│            [REVISE] ─→ Output v2               │
│                ↓                               │
│            [CRITIQUE] ─→ Feedback              │
│                ↓                               │
│           Score >= threshold? ─→ ...           │
└──────────────────────────────────────────────┘
```

### Reflection Prompt Templates

```
Generator prompt:
  "Write [output type] for [task]. Be thorough and accurate."

Critic prompt:
  "Review this [output type] against these criteria:
   1. [Criterion 1] (score 1-5)
   2. [Criterion 2] (score 1-5)
   ...
   Provide specific, actionable feedback for improvement.
   Overall score: X/5. If >= 4.5, mark as DONE."

Reviser prompt:
  "Revise this [output type] based on this feedback:
   [Feedback from critic]
   Keep what's already good. Only fix the identified issues.
   Do not add unnecessary caveats about the revision process."
```

### Reflection Best Practices

```
1. Use different temperatures:
   Generator: 0.3-0.7 (some creativity)
   Critic: 0 (deterministic evaluation)
   Reviser: 0.1-0.3 (focused changes)

2. Set a max iterations limit (2-3 usually sufficient)

3. Use specific, measurable criteria for the critic
   Bad: "Is it good?"
   Good: "Are all dates in ISO format? Are all claims cited?"

4. The critic should score each criterion separately
   - This prevents the "everything is fine" problem
   - It guides the reviser to focus on specific issues

5. Consider using a different model for critique
   - Stronger model as critic catches more issues
   - Same model may have blind spots about its own outputs
```

---

## Tool Calling Patterns

### OpenAI Function Calling

```python
# Define tools
tools = [{
    "type": "function",
    "function": {
        "name": "get_weather",
        "description": "Get the current weather for a location",
        "parameters": {
            "type": "object",
            "properties": {
                "location": {
                    "type": "string",
                    "description": "City and state, e.g., 'San Francisco, CA'"
                },
                "unit": {
                    "type": "string",
                    "enum": ["celsius", "fahrenheit"],
                    "default": "fahrenheit"
                }
            },
            "required": ["location"]
        }
    }
}]

# Call with tools
response = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "What's the weather in SF?"}],
    tools=tools,
    tool_choice="auto"  # or "required" to force tool use, or {"type": "function", "function": {"name": "get_weather"}}
)

# Handle tool calls
if response.choices[0].message.tool_calls:
    for tool_call in response.choices[0].message.tool_calls:
        args = json.loads(tool_call.function.arguments)
        result = execute_function(tool_call.function.name, args)

        # Send result back
        messages.append(response.choices[0].message)
        messages.append({
            "role": "tool",
            "tool_call_id": tool_call.id,
            "content": json.dumps(result)
        })
```

### Anthropic Tool Use

```python
# Define tools (Anthropic format)
tools = [{
    "name": "get_weather",
    "description": "Get the current weather for a location",
    "input_schema": {
        "type": "object",
        "properties": {
            "location": {
                "type": "string",
                "description": "City and state, e.g., 'San Francisco, CA'"
            }
        },
        "required": ["location"]
    }
}]

# Call with tools
response = client.messages.create(
    model="claude-sonnet-4-20250514",
    max_tokens=1024,
    tools=tools,
    messages=[{"role": "user", "content": "What's the weather in SF?"}]
)

# Handle tool use
for block in response.content:
    if block.type == "tool_use":
        result = execute_function(block.name, block.input)
        messages.append({"role": "assistant", "content": response.content})
        messages.append({
            "role": "user",
            "content": [{
                "type": "tool_result",
                "tool_use_id": block.id,
                "content": json.dumps(result)
            }]
        })
```

### Tool Design Guidelines

```
1. Names should be verb_noun:
   Good: search_products, create_ticket, send_email
   Bad: products, ticket, email

2. Descriptions should explain WHEN to use the tool:
   Good: "Search the product catalog. Use when the user asks about products,
          pricing, or availability."
   Bad: "Search products."

3. Parameters should have descriptions and constraints:
   Good: {"price": {"type": "number", "description": "Price in USD", "minimum": 0}}
   Bad: {"price": {"type": "number"}}

4. Use enums for finite choices:
   {"status": {"type": "string", "enum": ["open", "closed", "pending"]}}

5. Make tools idempotent where possible:
   Calling the same tool twice with same args should be safe

6. Return structured data:
   JSON with consistent schema, not free-form text

7. Include error information in tool results:
   {"success": false, "error": "Product not found", "suggestions": [...]}

8. Limit tool count:
   5-10 tools is optimal
   20+ tools degrades tool selection accuracy
   Group related operations into a single tool with a "action" parameter
```

---

## MCP (Model Context Protocol)

### Overview

```
MCP is an open protocol for connecting AI models to external data and tools.

Architecture:
  Client (Claude, ChatGPT, etc.) ←→ MCP Server ←→ External System

Key concepts:
  - Tools: Functions the model can call
  - Resources: Data the model can read
  - Prompts: Pre-built prompt templates
  - Sampling: Server-initiated LLM calls

Transport:
  - stdio: For local processes
  - HTTP+SSE: For remote servers
```

### MCP Server Structure

```python
# Minimal MCP server
from mcp.server import Server
from mcp.types import Tool, Resource

server = Server("my-server")

@server.tool("search")
async def search(query: str, limit: int = 10) -> str:
    """Search documents by keyword."""
    results = await db.search(query, limit)
    return json.dumps(results)

@server.resource("data://schema")
async def get_schema() -> str:
    """Get the database schema."""
    return await db.get_schema()

@server.prompt("analyze")
async def analyze_prompt(topic: str) -> list[dict]:
    """Generate analysis prompt for a topic."""
    return [
        {"role": "system", "content": f"You are analyzing {topic}."},
        {"role": "user", "content": f"Provide a detailed analysis of {topic}."}
    ]
```

### MCP vs Direct Tool Calling

```
MCP advantages:
  - Standardized protocol (works with any MCP-compatible client)
  - Resource discovery (client can list available tools/resources)
  - Bi-directional (server can request LLM calls)
  - Composable (multiple MCP servers can be connected)

Direct tool calling advantages:
  - Simpler (no protocol overhead)
  - Lower latency (no serialization/transport)
  - More flexible (any function signature)
  - Works offline

Use MCP when:
  - Building reusable integrations
  - Need to share tools across different AI clients
  - Want standardized discovery and documentation
  - Building a marketplace of tools

Use direct tool calling when:
  - Building a single application
  - Need maximum performance
  - Simple integration (1-5 tools)
  - All tools are internal
```

---

## Multi-Agent Patterns

### Orchestrator Pattern

```
┌──────────────────────────────────────────┐
│            ORCHESTRATOR                    │
│                                            │
│  Query → Decompose → Assign → Collect     │
│                ↓         ↓         ↓       │
│           ┌────────┐ ┌────────┐ ┌────────┐│
│           │Agent A │ │Agent B │ │Agent C ││
│           │Research│ │Code    │ │Review  ││
│           └────────┘ └────────┘ └────────┘│
│                ↓         ↓         ↓       │
│           Synthesize → Final Answer        │
└──────────────────────────────────────────┘

When to use: Tasks with clear sub-components that can be delegated
Example: "Build a feature" → Research Agent + Code Agent + Review Agent
```

### Pipeline Pattern

```
┌───────┐    ┌───────┐    ┌───────┐    ┌───────┐
│Agent 1│ →  │Agent 2│ →  │Agent 3│ →  │Agent 4│
│Extract│    │Analyze│    │Generate│    │Review │
└───────┘    └───────┘    └───────┘    └───────┘

When to use: Sequential processing where each stage transforms output
Example: Data extraction → Analysis → Report generation → Quality review
```

### Debate Pattern

```
┌──────────────────────────────────────────┐
│            DEBATE                          │
│                                            │
│  Query → Independent answers from all     │
│           ↓                                │
│  Round 1: Each agent sees others' answers │
│           ↓ Revise                         │
│  Round 2: Each agent sees others' answers │
│           ↓ Revise                         │
│  Judge: Select best or synthesize          │
└──────────────────────────────────────────┘

When to use: When you want to reduce individual model biases
Example: Contentious analysis, complex reasoning, fact-checking
Rounds: 2-3 is usually sufficient
```

### Supervisor Pattern

```
┌──────────────────────────────────────────┐
│           SUPERVISOR                       │
│                                            │
│  Monitors all worker agents                │
│  Decides next action based on state        │
│  Can: assign work, re-route, terminate     │
│                                            │
│  ┌────────┐ ┌────────┐ ┌────────┐        │
│  │Worker A│ │Worker B│ │Worker C│        │
│  └────────┘ └────────┘ └────────┘        │
│                                            │
│  Rules:                                    │
│  - If Worker A fails → reassign to B       │
│  - If quality < threshold → send to review │
│  - If all workers done → synthesize        │
└──────────────────────────────────────────┘

When to use: Need dynamic routing and error recovery
Example: Customer service with specialist handoff
```

---

## Agent Memory Patterns

### Memory Type Selection

| Memory Type | Persistence | Speed | Cost | Best For |
|------------|-------------|-------|------|----------|
| Buffer | Session only | Fast | Free | Simple conversations |
| Summary | Session only | Fast | LLM call/compress | Long conversations |
| Vector | Persistent | Medium | Embedding + storage | Long-term knowledge |
| Entity | Session/Persistent | Medium | LLM call/extract | User/product tracking |
| Knowledge Graph | Persistent | Medium | Complex | Relational queries |

### Memory Selection Guide

```
Simple chatbot: Buffer (last 10 messages)
Customer support: Buffer + Entity (track user details)
Research assistant: Buffer + Vector (remember findings)
Personal assistant: Buffer + Summary + Vector + Entity (full memory)
Enterprise agent: Vector + Knowledge Graph (organizational memory)
```

---

## Agent Safety Patterns

### Defense in Depth

```
Layer 1: Input validation (regex, length limits, encoding checks)
Layer 2: Prompt injection detection (classifier or pattern matching)
Layer 3: Tool permission system (which tools can the agent call?)
Layer 4: Output validation (PII check, content policy)
Layer 5: Human-in-the-loop (approval for high-risk actions)
Layer 6: Audit logging (record all agent actions)
Layer 7: Rate limiting (per-user, per-agent)
```

### Action Classification

```
Low risk (auto-approve):
  - Read operations (search, fetch, list)
  - In-memory computations
  - Generating text responses

Medium risk (log + execute):
  - Writing to databases
  - Calling external APIs
  - Creating files

High risk (require approval):
  - Deleting data
  - Sending emails/messages
  - Making purchases
  - Modifying permissions
  - Accessing sensitive data

Critical (always require human approval):
  - Financial transactions above threshold
  - Administrative actions
  - Irreversible operations
  - Actions affecting many users
```

### Agent Timeout and Resource Limits

```
Max steps per run: 10-20 (prevents infinite loops)
Max tool calls per step: 5 (prevents tool spam)
Max total tokens per run: 100K (prevents cost blowup)
Max wall-clock time: 5 minutes (prevents hanging)
Max concurrent tool calls: 3 (prevents resource exhaustion)
Max retry attempts per tool: 2 (prevents retry storms)

On any limit exceeded:
  1. Stop the agent
  2. Return partial results if available
  3. Log the limit hit for debugging
  4. Return a clear error to the user
```
