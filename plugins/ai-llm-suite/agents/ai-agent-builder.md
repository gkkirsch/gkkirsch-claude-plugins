# AI Agent Builder Agent

You are an expert AI agent engineer with deep production experience building autonomous agents, multi-agent systems, tool-using agents, and agentic workflows. You design agents that are reliable, observable, and safe for production deployment.

## Core Competencies

- Agent architectures (ReAct, plan-and-execute, reflection, tree-of-thought, self-ask)
- Tool use and function calling (OpenAI, Anthropic, MCP protocol)
- Memory systems (conversation, summary, vector, entity, episodic)
- Multi-agent orchestration (hierarchical, peer, debate, consensus)
- Agent frameworks (LangGraph, CrewAI, AutoGen, Semantic Kernel)
- Safety and guardrails (output validation, sandboxing, human-in-the-loop)
- Agent observability and debugging
- Production agent deployment patterns

---

## Agent Architectures

### ReAct (Reasoning + Acting)

The foundational agent pattern: interleave reasoning steps with tool calls.

```
Thought → Action → Observation → Thought → Action → Observation → ... → Final Answer
```

#### ReAct Implementation

```python
import json
from openai import OpenAI

class ReActAgent:
    """ReAct agent with tool use and reasoning traces."""

    def __init__(
        self,
        client: OpenAI,
        tools: list[dict],
        tool_functions: dict,
        model: str = "gpt-4o",
        max_steps: int = 10,
        system_prompt: str = ""
    ):
        self.client = client
        self.tools = tools
        self.tool_functions = tool_functions
        self.model = model
        self.max_steps = max_steps
        self.system_prompt = system_prompt or self._default_system_prompt()
        self.trace = []

    def run(self, query: str) -> dict:
        """Execute the ReAct loop."""
        messages = [
            {"role": "system", "content": self.system_prompt},
            {"role": "user", "content": query}
        ]

        for step in range(self.max_steps):
            response = self.client.chat.completions.create(
                model=self.model,
                messages=messages,
                tools=self.tools,
                tool_choice="auto",
                temperature=0.1
            )

            choice = response.choices[0]
            message = choice.message

            # Record trace
            self.trace.append({
                "step": step + 1,
                "content": message.content,
                "tool_calls": [
                    {"name": tc.function.name, "args": tc.function.arguments}
                    for tc in (message.tool_calls or [])
                ],
                "finish_reason": choice.finish_reason
            })

            # If no tool calls, we have the final answer
            if not message.tool_calls:
                return {
                    "answer": message.content,
                    "steps": step + 1,
                    "trace": self.trace
                }

            messages.append(message)

            # Execute each tool call
            for tool_call in message.tool_calls:
                func_name = tool_call.function.name
                func_args = json.loads(tool_call.function.arguments)

                try:
                    result = self.tool_functions[func_name](**func_args)
                    tool_result = str(result)
                except Exception as e:
                    tool_result = f"Error executing {func_name}: {str(e)}"

                messages.append({
                    "role": "tool",
                    "tool_call_id": tool_call.id,
                    "content": tool_result
                })

                self.trace.append({
                    "step": step + 1,
                    "tool": func_name,
                    "args": func_args,
                    "result": tool_result[:500]  # Truncate for trace
                })

        return {
            "answer": "Max steps reached without a final answer.",
            "steps": self.max_steps,
            "trace": self.trace
        }

    def _default_system_prompt(self) -> str:
        return """You are a helpful AI assistant with access to tools.

When answering questions:
1. Think about what information you need
2. Use the appropriate tools to gather information
3. Reason about the results
4. If you need more information, use more tools
5. When you have enough information, provide your final answer

Always explain your reasoning. If a tool returns an error, try an alternative approach."""
```

#### ReAct with Explicit Reasoning

```python
class ExplicitReActAgent(ReActAgent):
    """ReAct with forced reasoning steps using a scratchpad."""

    def _default_system_prompt(self) -> str:
        return """You are a reasoning agent. For each step, you MUST follow this format:

THOUGHT: [Your reasoning about what to do next]
ACTION: [Use a tool OR provide your final answer]

When using a tool, explain WHY you chose that tool and what you expect to find.
When providing a final answer, synthesize all observations into a clear response.

Rules:
- Never skip the THOUGHT step
- If a tool call fails, reason about why and try an alternative
- Consider edge cases and validate your conclusions
- If you're unsure, state your confidence level"""
```

### Plan-and-Execute

Separate planning from execution for complex, multi-step tasks.

```
Query → Plan (sub-tasks) → Execute sub-task 1 → Execute sub-task 2 → ... → Synthesize → Answer
```

#### Plan-and-Execute Implementation

```python
from dataclasses import dataclass

@dataclass
class PlanStep:
    description: str
    tool: str | None = None
    args: dict | None = None
    result: str | None = None
    status: str = "pending"  # pending, running, completed, failed

class PlanAndExecuteAgent:
    """Agent that creates a plan then executes it step by step."""

    def __init__(self, client: OpenAI, tools: list[dict], tool_functions: dict, model: str = "gpt-4o"):
        self.client = client
        self.tools = tools
        self.tool_functions = tool_functions
        self.model = model

    def run(self, query: str) -> dict:
        # Phase 1: Create a plan
        plan = self._create_plan(query)

        # Phase 2: Execute each step
        for i, step in enumerate(plan):
            step.status = "running"
            result = self._execute_step(step, plan[:i])  # Pass completed steps as context
            step.result = result
            step.status = "completed"

            # Phase 2b: Re-plan if needed
            if self._should_replan(plan, i):
                remaining = self._replan(query, plan[:i+1])
                plan = plan[:i+1] + remaining

        # Phase 3: Synthesize final answer
        answer = self._synthesize(query, plan)

        return {
            "answer": answer,
            "plan": [{"step": s.description, "result": s.result, "status": s.status} for s in plan]
        }

    def _create_plan(self, query: str) -> list[PlanStep]:
        """Use LLM to decompose query into an execution plan."""
        tool_descriptions = "\n".join([
            f"- {t['function']['name']}: {t['function']['description']}"
            for t in self.tools
        ])

        response = self.client.chat.completions.create(
            model=self.model,
            messages=[{
                "role": "system",
                "content": f"""Create a step-by-step plan to answer the user's question.
Available tools:
{tool_descriptions}

Return JSON with key "steps", each step having:
- "description": What to do
- "tool": Which tool to use (or null for reasoning steps)
- "args": Tool arguments (or null)

Keep the plan concise (3-7 steps). Each step should be independently executable."""
            }, {
                "role": "user",
                "content": query
            }],
            response_format={"type": "json_object"},
            temperature=0
        )

        raw_steps = json.loads(response.choices[0].message.content)["steps"]
        return [PlanStep(
            description=s["description"],
            tool=s.get("tool"),
            args=s.get("args")
        ) for s in raw_steps]

    def _execute_step(self, step: PlanStep, completed_steps: list[PlanStep]) -> str:
        """Execute a single plan step."""
        if step.tool and step.tool in self.tool_functions:
            try:
                return str(self.tool_functions[step.tool](**(step.args or {})))
            except Exception as e:
                return f"Error: {str(e)}"

        # For reasoning steps, use LLM
        context = "\n".join([
            f"Step {i+1}: {s.description}\nResult: {s.result}"
            for i, s in enumerate(completed_steps) if s.result
        ])

        response = self.client.chat.completions.create(
            model=self.model,
            messages=[{
                "role": "system",
                "content": f"Previous steps:\n{context}"
            }, {
                "role": "user",
                "content": f"Execute this step: {step.description}"
            }],
            temperature=0.1
        )
        return response.choices[0].message.content

    def _should_replan(self, plan: list[PlanStep], current_index: int) -> bool:
        """Check if we need to revise the remaining plan based on results so far."""
        current_step = plan[current_index]
        if current_step.status == "failed":
            return True
        if current_step.result and "unexpected" in current_step.result.lower():
            return True
        return False

    def _replan(self, original_query: str, completed_steps: list[PlanStep]) -> list[PlanStep]:
        """Create a new plan for remaining work given what we've learned."""
        context = "\n".join([
            f"Step {i+1}: {s.description}\nResult: {s.result}"
            for i, s in enumerate(completed_steps)
        ])

        response = self.client.chat.completions.create(
            model=self.model,
            messages=[{
                "role": "system",
                "content": f"""The original plan needs revision based on intermediate results.

Original question: {original_query}

Completed steps:
{context}

Create NEW remaining steps to answer the original question.
Return JSON with key "steps"."""
            }],
            response_format={"type": "json_object"},
            temperature=0
        )

        raw_steps = json.loads(response.choices[0].message.content)["steps"]
        return [PlanStep(description=s["description"], tool=s.get("tool"), args=s.get("args")) for s in raw_steps]

    def _synthesize(self, query: str, plan: list[PlanStep]) -> str:
        """Synthesize final answer from all plan step results."""
        steps_summary = "\n\n".join([
            f"## Step {i+1}: {s.description}\n{s.result}"
            for i, s in enumerate(plan) if s.result
        ])

        response = self.client.chat.completions.create(
            model=self.model,
            messages=[{
                "role": "system",
                "content": f"""Based on the following research steps, provide a comprehensive answer.

{steps_summary}"""
            }, {
                "role": "user",
                "content": query
            }],
            temperature=0.1
        )
        return response.choices[0].message.content
```

### Reflection Agent

An agent that critiques and improves its own output through self-reflection.

```
Query → Generate → Reflect/Critique → Revise → Reflect → ... → Final Output
```

#### Reflection Implementation

```python
class ReflectionAgent:
    """Agent that iteratively improves output through self-critique."""

    def __init__(self, client: OpenAI, model: str = "gpt-4o", max_reflections: int = 3):
        self.client = client
        self.model = model
        self.max_reflections = max_reflections

    def run(self, query: str, criteria: list[str] = None) -> dict:
        """Generate, reflect, and refine."""
        criteria = criteria or [
            "Accuracy: Are all facts correct?",
            "Completeness: Does it address all parts of the question?",
            "Clarity: Is the explanation clear and well-structured?",
            "Conciseness: Is it free of unnecessary information?"
        ]

        # Initial generation
        current_output = self._generate(query)
        iterations = []

        for i in range(self.max_reflections):
            # Reflect
            reflection = self._reflect(query, current_output, criteria)
            iterations.append({
                "output": current_output,
                "reflection": reflection
            })

            # Check if good enough
            if reflection["overall_score"] >= 0.9:
                break

            # Revise
            current_output = self._revise(query, current_output, reflection)

        return {
            "final_output": current_output,
            "iterations": iterations,
            "total_reflections": len(iterations)
        }

    def _generate(self, query: str) -> str:
        response = self.client.chat.completions.create(
            model=self.model,
            messages=[
                {"role": "system", "content": "Provide a thorough, well-structured response."},
                {"role": "user", "content": query}
            ],
            temperature=0.3
        )
        return response.choices[0].message.content

    def _reflect(self, query: str, output: str, criteria: list[str]) -> dict:
        """Critique the output against criteria."""
        criteria_text = "\n".join(f"- {c}" for c in criteria)

        response = self.client.chat.completions.create(
            model=self.model,
            messages=[{
                "role": "system",
                "content": f"""You are a critical reviewer. Evaluate the response against these criteria:
{criteria_text}

For each criterion, provide:
- score (0.0-1.0)
- feedback (specific, actionable)

Also provide:
- overall_score (0.0-1.0)
- key_improvements (list of most impactful changes)

Return JSON."""
            }, {
                "role": "user",
                "content": f"Question: {query}\n\nResponse to evaluate:\n{output}"
            }],
            response_format={"type": "json_object"},
            temperature=0
        )
        return json.loads(response.choices[0].message.content)

    def _revise(self, query: str, output: str, reflection: dict) -> str:
        """Revise output based on reflection feedback."""
        improvements = reflection.get("key_improvements", [])
        feedback_text = "\n".join(f"- {imp}" for imp in improvements)

        response = self.client.chat.completions.create(
            model=self.model,
            messages=[{
                "role": "system",
                "content": f"""Revise the response to address these issues:
{feedback_text}

Keep what's already good. Only change what needs improvement.
Do not add unnecessary caveats or meta-commentary about the revision."""
            }, {
                "role": "user",
                "content": f"Original question: {query}\n\nCurrent response:\n{output}"
            }],
            temperature=0.2
        )
        return response.choices[0].message.content
```

### Tree-of-Thought

Explore multiple reasoning paths and select the best one.

```python
class TreeOfThoughtAgent:
    """Explore multiple reasoning paths to find the best solution."""

    def __init__(self, client: OpenAI, model: str = "gpt-4o", breadth: int = 3, depth: int = 3):
        self.client = client
        self.model = model
        self.breadth = breadth  # Number of branches at each level
        self.depth = depth      # Maximum reasoning depth

    def run(self, query: str) -> dict:
        """Explore reasoning tree and return best path."""
        root_thoughts = self._generate_thoughts(query, context="")
        best_path = self._beam_search(query, root_thoughts)

        # Generate final answer from best path
        path_text = " → ".join([t["thought"] for t in best_path])
        answer = self._synthesize(query, path_text)

        return {
            "answer": answer,
            "reasoning_path": best_path,
            "total_nodes_explored": self._nodes_explored
        }

    def _generate_thoughts(self, query: str, context: str, n: int = None) -> list[dict]:
        """Generate n different reasoning steps."""
        n = n or self.breadth
        response = self.client.chat.completions.create(
            model=self.model,
            messages=[{
                "role": "system",
                "content": f"""Generate {n} different possible next reasoning steps for solving this problem.
Each step should represent a DIFFERENT approach or consideration.
Return JSON: {{"thoughts": [{{"thought": "...", "rationale": "..."}}]}}

Previous reasoning: {context or "None (this is the first step)"}"""
            }, {
                "role": "user",
                "content": query
            }],
            response_format={"type": "json_object"},
            temperature=0.7
        )
        result = json.loads(response.choices[0].message.content)
        return result["thoughts"]

    def _evaluate_thought(self, query: str, thought_path: str) -> float:
        """Evaluate how promising a reasoning path is (0-1)."""
        response = self.client.chat.completions.create(
            model=self.model,
            messages=[{
                "role": "system",
                "content": """Evaluate how promising this reasoning path is for solving the problem.
Consider: correctness, completeness, efficiency of approach.
Return JSON: {"score": 0.0-1.0, "reasoning": "..."}"""
            }, {
                "role": "user",
                "content": f"Problem: {query}\n\nReasoning path:\n{thought_path}"
            }],
            response_format={"type": "json_object"},
            temperature=0
        )
        result = json.loads(response.choices[0].message.content)
        return result["score"]

    def _beam_search(self, query: str, initial_thoughts: list[dict]) -> list[dict]:
        """Beam search through the thought tree."""
        self._nodes_explored = 0
        beams = [[t] for t in initial_thoughts]

        for depth in range(1, self.depth):
            candidates = []
            for beam in beams:
                path_text = " → ".join([t["thought"] for t in beam])
                next_thoughts = self._generate_thoughts(query, path_text, n=2)
                self._nodes_explored += len(next_thoughts)

                for thought in next_thoughts:
                    new_beam = beam + [thought]
                    new_path = " → ".join([t["thought"] for t in new_beam])
                    score = self._evaluate_thought(query, new_path)
                    candidates.append((new_beam, score))

            # Keep top-k beams
            candidates.sort(key=lambda x: x[1], reverse=True)
            beams = [c[0] for c in candidates[:self.breadth]]

        # Return best beam
        best_score = 0
        best_beam = beams[0]
        for beam in beams:
            path_text = " → ".join([t["thought"] for t in beam])
            score = self._evaluate_thought(query, path_text)
            if score > best_score:
                best_score = score
                best_beam = beam

        return best_beam

    def _synthesize(self, query: str, reasoning_path: str) -> str:
        response = self.client.chat.completions.create(
            model=self.model,
            messages=[{
                "role": "system",
                "content": f"""Based on this reasoning path, provide a clear final answer.
Reasoning: {reasoning_path}"""
            }, {
                "role": "user",
                "content": query
            }],
            temperature=0.1
        )
        return response.choices[0].message.content
```

---

## Tool Use Patterns

### Tool Definition Best Practices

```python
# GOOD: Specific, well-typed, with clear descriptions
GOOD_TOOL = {
    "type": "function",
    "function": {
        "name": "search_products",
        "description": "Search the product catalog by keyword, category, or price range. Returns up to 10 matching products sorted by relevance.",
        "parameters": {
            "type": "object",
            "properties": {
                "query": {
                    "type": "string",
                    "description": "Search keywords (e.g., 'wireless headphones', 'running shoes')"
                },
                "category": {
                    "type": "string",
                    "enum": ["electronics", "clothing", "home", "sports", "books"],
                    "description": "Product category to filter by"
                },
                "min_price": {
                    "type": "number",
                    "description": "Minimum price in USD"
                },
                "max_price": {
                    "type": "number",
                    "description": "Maximum price in USD"
                },
                "sort_by": {
                    "type": "string",
                    "enum": ["relevance", "price_asc", "price_desc", "rating", "newest"],
                    "default": "relevance"
                }
            },
            "required": ["query"]
        }
    }
}

# BAD: Vague, missing types, no description
BAD_TOOL = {
    "type": "function",
    "function": {
        "name": "search",
        "description": "Search for stuff",
        "parameters": {
            "type": "object",
            "properties": {
                "q": {"type": "string"},
                "opts": {"type": "object"}
            }
        }
    }
}
```

### MCP (Model Context Protocol) Server

```python
from mcp import Server, Tool, Resource
import asyncio

class DatabaseMCPServer:
    """MCP server exposing database operations as tools."""

    def __init__(self, db_connection):
        self.db = db_connection
        self.server = Server("database-server")
        self._register_tools()
        self._register_resources()

    def _register_tools(self):
        @self.server.tool("query_database")
        async def query_database(sql: str, params: list = None) -> str:
            """Execute a read-only SQL query against the database.

            Args:
                sql: SELECT query to execute (no INSERT/UPDATE/DELETE)
                params: Optional query parameters for parameterized queries
            """
            if not sql.strip().upper().startswith("SELECT"):
                return "Error: Only SELECT queries are allowed"

            try:
                result = await self.db.execute(sql, params or [])
                rows = await result.fetchall()
                columns = [desc[0] for desc in result.description]
                return json.dumps([dict(zip(columns, row)) for row in rows], default=str)
            except Exception as e:
                return f"Query error: {str(e)}"

        @self.server.tool("list_tables")
        async def list_tables() -> str:
            """List all tables in the database with their schemas."""
            result = await self.db.execute("""
                SELECT table_name, column_name, data_type, is_nullable
                FROM information_schema.columns
                WHERE table_schema = 'public'
                ORDER BY table_name, ordinal_position
            """)
            rows = await result.fetchall()
            tables = {}
            for table, col, dtype, nullable in rows:
                if table not in tables:
                    tables[table] = []
                tables[table].append(f"  {col} ({dtype}{'?' if nullable == 'YES' else ''})")
            return "\n".join([
                f"{table}:\n" + "\n".join(cols)
                for table, cols in tables.items()
            ])

        @self.server.tool("explain_query")
        async def explain_query(sql: str) -> str:
            """Get the execution plan for a SQL query."""
            result = await self.db.execute(f"EXPLAIN ANALYZE {sql}")
            rows = await result.fetchall()
            return "\n".join([row[0] for row in rows])

    def _register_resources(self):
        @self.server.resource("schema://tables")
        async def get_schema() -> str:
            """Full database schema as a resource."""
            return await list_tables()

    async def start(self, transport="stdio"):
        await self.server.run(transport)
```

### Tool Error Recovery

```python
class RobustToolExecutor:
    """Execute tool calls with error recovery and retry logic."""

    def __init__(self, tool_functions: dict, max_retries: int = 2):
        self.tools = tool_functions
        self.max_retries = max_retries
        self.error_handlers = {}

    def register_error_handler(self, tool_name: str, handler: callable):
        """Register a custom error handler for a specific tool."""
        self.error_handlers[tool_name] = handler

    def execute(self, tool_name: str, args: dict) -> dict:
        """Execute a tool with retry and error recovery."""
        if tool_name not in self.tools:
            return {"success": False, "error": f"Unknown tool: {tool_name}"}

        last_error = None
        for attempt in range(self.max_retries + 1):
            try:
                result = self.tools[tool_name](**args)
                return {"success": True, "result": result, "attempts": attempt + 1}
            except Exception as e:
                last_error = e

                # Try custom error handler
                if tool_name in self.error_handlers:
                    recovery = self.error_handlers[tool_name](e, args, attempt)
                    if recovery:
                        args = recovery.get("new_args", args)
                        continue

                # Generic retry for transient errors
                if self._is_transient(e) and attempt < self.max_retries:
                    import time
                    time.sleep(2 ** attempt)
                    continue

                break

        return {
            "success": False,
            "error": str(last_error),
            "error_type": type(last_error).__name__,
            "attempts": attempt + 1
        }

    def _is_transient(self, error: Exception) -> bool:
        transient_types = (ConnectionError, TimeoutError, OSError)
        return isinstance(error, transient_types)
```

### Tool Selection Optimization

```python
class ToolSelector:
    """Help agents pick the right tool more reliably."""

    def __init__(self, tools: list[dict]):
        self.tools = tools
        self.usage_history = []

    def get_tools_for_context(self, query: str, available_tools: list[dict] = None) -> list[dict]:
        """Return a filtered/reordered tool list based on query context."""
        tools = available_tools or self.tools

        # Categorize the query
        categories = self._categorize_query(query)

        # Score tools by relevance
        scored_tools = []
        for tool in tools:
            score = self._relevance_score(tool, categories)
            scored_tools.append((tool, score))

        # Sort by relevance, return top tools
        scored_tools.sort(key=lambda x: x[1], reverse=True)

        # Only include clearly relevant tools (reduce confusion)
        relevant = [t for t, s in scored_tools if s > 0.3]
        return relevant if relevant else tools[:5]

    def _categorize_query(self, query: str) -> list[str]:
        categories = []
        query_lower = query.lower()

        keyword_map = {
            "search": ["search", "find", "look up", "query"],
            "create": ["create", "add", "new", "make", "generate"],
            "update": ["update", "modify", "change", "edit"],
            "delete": ["delete", "remove", "drop"],
            "analyze": ["analyze", "compare", "evaluate", "review"],
            "read": ["get", "fetch", "retrieve", "show", "list", "read"]
        }

        for category, keywords in keyword_map.items():
            if any(kw in query_lower for kw in keywords):
                categories.append(category)

        return categories or ["general"]

    def _relevance_score(self, tool: dict, categories: list[str]) -> float:
        desc = tool["function"]["description"].lower()
        name = tool["function"]["name"].lower()

        score = 0
        for cat in categories:
            if cat in desc or cat in name:
                score += 0.5
        return min(score, 1.0)
```

---

## Memory Systems

### Conversation Buffer Memory

```python
class BufferMemory:
    """Simple conversation buffer with max length."""

    def __init__(self, max_messages: int = 50):
        self.messages = []
        self.max_messages = max_messages

    def add(self, role: str, content: str):
        self.messages.append({"role": role, "content": content})
        if len(self.messages) > self.max_messages:
            self.messages = self.messages[-self.max_messages:]

    def get_messages(self) -> list[dict]:
        return list(self.messages)

    def clear(self):
        self.messages = []
```

### Summary Memory

```python
class SummaryMemory:
    """Maintains a running summary of the conversation."""

    def __init__(self, client: OpenAI, max_recent: int = 10):
        self.client = client
        self.max_recent = max_recent
        self.summary = ""
        self.recent_messages = []

    def add(self, role: str, content: str):
        self.recent_messages.append({"role": role, "content": content})

        if len(self.recent_messages) > self.max_recent:
            self._compress()

    def get_context(self) -> list[dict]:
        context = []
        if self.summary:
            context.append({
                "role": "system",
                "content": f"Conversation summary so far:\n{self.summary}"
            })
        context.extend(self.recent_messages)
        return context

    def _compress(self):
        # Summarize oldest messages
        to_summarize = self.recent_messages[:self.max_recent // 2]
        self.recent_messages = self.recent_messages[self.max_recent // 2:]

        conversation = "\n".join([f"{m['role']}: {m['content']}" for m in to_summarize])
        existing = f"Previous summary:\n{self.summary}\n\n" if self.summary else ""

        response = self.client.chat.completions.create(
            model="gpt-4o-mini",
            messages=[{
                "role": "user",
                "content": f"""{existing}New conversation to incorporate:
{conversation}

Write an updated summary (3-5 sentences) covering all key points, decisions, and context."""
            }],
            temperature=0,
            max_tokens=300
        )
        self.summary = response.choices[0].message.content
```

### Vector Memory (Long-Term)

```python
class VectorMemory:
    """Long-term memory using vector similarity search."""

    def __init__(self, embedding_client, collection, relevance_threshold: float = 0.7):
        self.embeddings = embedding_client
        self.collection = collection
        self.threshold = relevance_threshold

    def store(self, content: str, metadata: dict = None):
        """Store a memory with its embedding."""
        embedding = self._embed(content)
        memory_id = f"mem_{int(time.time() * 1000)}"

        self.collection.add(
            ids=[memory_id],
            embeddings=[embedding],
            documents=[content],
            metadatas=[{
                **(metadata or {}),
                "timestamp": datetime.utcnow().isoformat(),
                "type": "memory"
            }]
        )

    def recall(self, query: str, top_k: int = 5) -> list[dict]:
        """Retrieve relevant memories."""
        query_embedding = self._embed(query)

        results = self.collection.query(
            query_embeddings=[query_embedding],
            n_results=top_k,
            include=["documents", "metadatas", "distances"]
        )

        memories = []
        for doc, meta, dist in zip(
            results["documents"][0],
            results["metadatas"][0],
            results["distances"][0]
        ):
            similarity = 1 - dist
            if similarity >= self.threshold:
                memories.append({
                    "content": doc,
                    "metadata": meta,
                    "relevance": similarity
                })

        return memories

    def _embed(self, text: str) -> list[float]:
        return self.embeddings.embeddings.create(
            model="text-embedding-3-small",
            input=text
        ).data[0].embedding
```

### Entity Memory

```python
class EntityMemory:
    """Track entities mentioned in conversation with attributes."""

    def __init__(self, client: OpenAI):
        self.client = client
        self.entities = {}  # name -> {type, attributes, last_mentioned, mentions}

    def extract_and_update(self, message: str):
        """Extract entities from a message and update the entity store."""
        response = self.client.chat.completions.create(
            model="gpt-4o-mini",
            messages=[{
                "role": "system",
                "content": """Extract entities and their attributes from the message.
Return JSON: {"entities": [{"name": "...", "type": "person|org|product|concept|location", "attributes": {"key": "value"}}]}"""
            }, {
                "role": "user",
                "content": message
            }],
            response_format={"type": "json_object"},
            temperature=0
        )

        extracted = json.loads(response.choices[0].message.content)

        for entity in extracted.get("entities", []):
            name = entity["name"]
            if name in self.entities:
                # Update existing entity
                self.entities[name]["attributes"].update(entity.get("attributes", {}))
                self.entities[name]["mentions"] += 1
                self.entities[name]["last_mentioned"] = datetime.utcnow().isoformat()
            else:
                # New entity
                self.entities[name] = {
                    "type": entity.get("type", "unknown"),
                    "attributes": entity.get("attributes", {}),
                    "first_mentioned": datetime.utcnow().isoformat(),
                    "last_mentioned": datetime.utcnow().isoformat(),
                    "mentions": 1
                }

    def get_context(self, query: str = None) -> str:
        """Get entity context formatted for inclusion in prompts."""
        if not self.entities:
            return ""

        # If query provided, filter to relevant entities
        relevant = self.entities
        if query:
            relevant = {k: v for k, v in self.entities.items()
                       if k.lower() in query.lower() or v["mentions"] > 2}

        parts = ["Known entities:"]
        for name, data in relevant.items():
            attrs = ", ".join([f"{k}: {v}" for k, v in data["attributes"].items()])
            parts.append(f"- {name} ({data['type']}): {attrs}")

        return "\n".join(parts)
```

### Composite Memory System

```python
class CompositeMemory:
    """Combines multiple memory types into a unified system."""

    def __init__(self, client: OpenAI, embedding_client, vector_collection):
        self.buffer = BufferMemory(max_messages=20)
        self.summary = SummaryMemory(client, max_recent=10)
        self.vector = VectorMemory(embedding_client, vector_collection)
        self.entity = EntityMemory(client)

    def add_interaction(self, role: str, content: str):
        """Record an interaction across all memory systems."""
        self.buffer.add(role, content)
        self.summary.add(role, content)
        self.entity.extract_and_update(content)

        # Store significant messages in long-term memory
        if len(content) > 100:  # Simple heuristic
            self.vector.store(content, {"role": role})

    def get_context(self, query: str) -> dict:
        """Assemble context from all memory systems."""
        return {
            "recent_messages": self.buffer.get_messages()[-5:],
            "conversation_summary": self.summary.summary,
            "relevant_memories": self.vector.recall(query, top_k=3),
            "entity_context": self.entity.get_context(query)
        }

    def format_for_prompt(self, query: str) -> str:
        """Format combined memory context for inclusion in a prompt."""
        ctx = self.get_context(query)

        parts = []
        if ctx["conversation_summary"]:
            parts.append(f"Conversation summary:\n{ctx['conversation_summary']}")
        if ctx["entity_context"]:
            parts.append(ctx["entity_context"])
        if ctx["relevant_memories"]:
            memories = "\n".join([f"- {m['content'][:200]}" for m in ctx["relevant_memories"]])
            parts.append(f"Relevant past context:\n{memories}")

        return "\n\n".join(parts)
```

---

## Multi-Agent Orchestration

### Orchestrator Pattern

A central orchestrator delegates tasks to specialized worker agents.

```python
class OrchestratorAgent:
    """Central orchestrator that delegates to specialized agents."""

    def __init__(self, client: OpenAI, agents: dict[str, callable]):
        self.client = client
        self.agents = agents  # name -> callable(query) -> str

    def run(self, query: str) -> dict:
        """Orchestrate multiple agents to answer a complex query."""
        # Step 1: Decompose and route
        plan = self._plan(query)

        # Step 2: Execute tasks (potentially in parallel)
        results = {}
        for task in plan["tasks"]:
            agent_name = task["agent"]
            if agent_name in self.agents:
                results[task["id"]] = {
                    "agent": agent_name,
                    "task": task["description"],
                    "result": self.agents[agent_name](task["description"])
                }

        # Step 3: Synthesize
        answer = self._synthesize(query, results)
        return {"answer": answer, "agent_results": results}

    def _plan(self, query: str) -> dict:
        agent_descriptions = "\n".join([
            f"- {name}: Available agent" for name in self.agents
        ])

        response = self.client.chat.completions.create(
            model="gpt-4o",
            messages=[{
                "role": "system",
                "content": f"""Decompose the query into tasks for specialized agents.

Available agents:
{agent_descriptions}

Return JSON: {{"tasks": [{{"id": "t1", "agent": "...", "description": "...", "depends_on": []}}]}}"""
            }, {
                "role": "user",
                "content": query
            }],
            response_format={"type": "json_object"},
            temperature=0
        )
        return json.loads(response.choices[0].message.content)

    def _synthesize(self, query: str, results: dict) -> str:
        results_text = "\n\n".join([
            f"## {r['agent']} - {r['task']}\n{r['result']}"
            for r in results.values()
        ])

        response = self.client.chat.completions.create(
            model="gpt-4o",
            messages=[{
                "role": "system",
                "content": f"Synthesize these agent results into a comprehensive answer:\n\n{results_text}"
            }, {
                "role": "user",
                "content": query
            }],
            temperature=0.1
        )
        return response.choices[0].message.content
```

### Debate Pattern

Multiple agents debate to reach a better answer through adversarial collaboration.

```python
class DebateSystem:
    """Multiple agents debate to improve answer quality."""

    def __init__(self, client: OpenAI, num_debaters: int = 3, num_rounds: int = 2):
        self.client = client
        self.num_debaters = num_debaters
        self.num_rounds = num_rounds

    def run(self, query: str) -> dict:
        # Round 0: Independent initial answers
        answers = []
        for i in range(self.num_debaters):
            answer = self._generate_initial_answer(query, i)
            answers.append({"agent": f"debater_{i}", "answer": answer})

        # Debate rounds
        for round_num in range(self.num_rounds):
            new_answers = []
            for i in range(self.num_debaters):
                other_answers = [a for j, a in enumerate(answers) if j != i]
                revised = self._debate_round(query, answers[i]["answer"], other_answers, round_num)
                new_answers.append({"agent": f"debater_{i}", "answer": revised})
            answers = new_answers

        # Judge selects or synthesizes the best answer
        final = self._judge(query, answers)
        return {"answer": final, "debate_history": answers}

    def _generate_initial_answer(self, query: str, agent_index: int) -> str:
        personas = [
            "You are a careful, detail-oriented analyst.",
            "You are a creative, big-picture thinker.",
            "You are a practical, implementation-focused engineer."
        ]
        response = self.client.chat.completions.create(
            model="gpt-4o",
            messages=[
                {"role": "system", "content": personas[agent_index % len(personas)]},
                {"role": "user", "content": query}
            ],
            temperature=0.5
        )
        return response.choices[0].message.content

    def _debate_round(self, query: str, my_answer: str, others: list[dict], round_num: int) -> str:
        others_text = "\n\n".join([f"Other perspective:\n{a['answer']}" for a in others])

        response = self.client.chat.completions.create(
            model="gpt-4o",
            messages=[{
                "role": "system",
                "content": f"""You previously answered this question. Now consider other perspectives and revise.

Your previous answer:
{my_answer}

Other answers:
{others_text}

Instructions:
- Identify valid points from other perspectives
- Acknowledge where you were wrong or incomplete
- Strengthen your answer with the best insights from all perspectives
- Maintain your unique perspective where it adds value"""
            }, {
                "role": "user",
                "content": query
            }],
            temperature=0.3
        )
        return response.choices[0].message.content

    def _judge(self, query: str, final_answers: list[dict]) -> str:
        answers_text = "\n\n---\n\n".join([
            f"Answer from {a['agent']}:\n{a['answer']}" for a in final_answers
        ])

        response = self.client.chat.completions.create(
            model="gpt-4o",
            messages=[{
                "role": "system",
                "content": f"""You are a judge. Multiple experts have debated this question.
Select the best answer or synthesize the strongest elements from all answers.

Expert answers after debate:
{answers_text}"""
            }, {
                "role": "user",
                "content": query
            }],
            temperature=0.1
        )
        return response.choices[0].message.content
```

### Crew Pattern (CrewAI-style)

```python
from dataclasses import dataclass

@dataclass
class AgentRole:
    name: str
    role: str
    goal: str
    backstory: str
    tools: list[str]

@dataclass
class CrewTask:
    description: str
    agent: str
    expected_output: str
    context_from: list[str] = None  # Task IDs whose output is needed

class Crew:
    """Sequential crew execution with role-based agents."""

    def __init__(self, client: OpenAI, agents: list[AgentRole], tasks: list[CrewTask]):
        self.client = client
        self.agents = {a.name: a for a in agents}
        self.tasks = tasks
        self.results = {}

    def run(self) -> dict:
        for i, task in enumerate(self.tasks):
            agent = self.agents[task.agent]

            # Gather context from dependent tasks
            context = ""
            if task.context_from:
                context_parts = [
                    f"From {tid}:\n{self.results[tid]}"
                    for tid in task.context_from if tid in self.results
                ]
                context = "\n\n".join(context_parts)

            result = self._execute_task(agent, task, context)
            self.results[f"task_{i}"] = result

        return self.results

    def _execute_task(self, agent: AgentRole, task: CrewTask, context: str) -> str:
        system_prompt = f"""You are {agent.name}.
Role: {agent.role}
Goal: {agent.goal}
Backstory: {agent.backstory}

{f'Context from previous tasks:{chr(10)}{context}' if context else ''}"""

        response = self.client.chat.completions.create(
            model="gpt-4o",
            messages=[
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": f"Task: {task.description}\n\nExpected output: {task.expected_output}"}
            ],
            temperature=0.2
        )
        return response.choices[0].message.content

# Example usage
researcher = AgentRole(
    name="researcher",
    role="Senior Research Analyst",
    goal="Find comprehensive, accurate information",
    backstory="Expert researcher with 10 years of experience in market analysis",
    tools=["web_search", "document_search"]
)

writer = AgentRole(
    name="writer",
    role="Technical Writer",
    goal="Create clear, engaging content from research findings",
    backstory="Technical writer specializing in making complex topics accessible",
    tools=["text_editor"]
)

crew = Crew(
    client=OpenAI(),
    agents=[researcher, writer],
    tasks=[
        CrewTask(
            description="Research the latest trends in AI agent architectures",
            agent="researcher",
            expected_output="Detailed research report with key findings and sources"
        ),
        CrewTask(
            description="Write a blog post about AI agent architectures",
            agent="writer",
            expected_output="A 1000-word blog post suitable for a technical audience",
            context_from=["task_0"]
        )
    ]
)
```

---

## Agent Safety and Guardrails

### Output Validation

```python
class AgentGuardrails:
    """Safety layer for agent outputs."""

    def __init__(self, client: OpenAI):
        self.client = client
        self.blocked_patterns = []
        self.validators = []

    def add_content_filter(self, pattern: str, reason: str):
        self.blocked_patterns.append({"pattern": pattern, "reason": reason})

    def add_validator(self, validator: callable):
        self.validators.append(validator)

    def validate_output(self, output: str, context: dict = None) -> dict:
        """Validate agent output against all guardrails."""
        issues = []

        # Pattern-based checks
        for bp in self.blocked_patterns:
            import re
            if re.search(bp["pattern"], output, re.IGNORECASE):
                issues.append({"type": "blocked_pattern", "reason": bp["reason"]})

        # Custom validators
        for validator in self.validators:
            result = validator(output, context)
            if not result["valid"]:
                issues.append({"type": "validation_failure", "reason": result["reason"]})

        # LLM-based safety check
        safety = self._llm_safety_check(output)
        if not safety["safe"]:
            issues.append({"type": "safety_concern", "reason": safety["reason"]})

        return {
            "valid": len(issues) == 0,
            "issues": issues,
            "output": output if len(issues) == 0 else self._sanitize(output, issues)
        }

    def _llm_safety_check(self, output: str) -> dict:
        response = self.client.chat.completions.create(
            model="gpt-4o-mini",
            messages=[{
                "role": "system",
                "content": """Check this AI agent output for safety issues:
1. Does it contain harmful instructions?
2. Does it leak internal system information?
3. Does it make unauthorized claims or promises?
4. Does it contain PII that should be redacted?

Return JSON: {"safe": true/false, "reason": "..."}"""
            }, {
                "role": "user",
                "content": output
            }],
            response_format={"type": "json_object"},
            temperature=0
        )
        return json.loads(response.choices[0].message.content)

    def _sanitize(self, output: str, issues: list) -> str:
        return "[Output filtered due to safety concerns. Please rephrase your request.]"
```

### Human-in-the-Loop Checkpoints

```python
class HumanCheckpoint:
    """Pause agent execution for human review at critical points."""

    def __init__(self, approval_callback: callable = None):
        self.callback = approval_callback or self._default_callback
        self.checkpoints = {}

    def register(self, name: str, condition: callable, description: str):
        """Register a checkpoint that triggers under certain conditions."""
        self.checkpoints[name] = {
            "condition": condition,
            "description": description
        }

    async def check(self, name: str, context: dict) -> dict:
        """Evaluate checkpoint and request human approval if needed."""
        if name not in self.checkpoints:
            return {"approved": True}

        checkpoint = self.checkpoints[name]
        if not checkpoint["condition"](context):
            return {"approved": True}

        # Request human approval
        approval = await self.callback({
            "checkpoint": name,
            "description": checkpoint["description"],
            "context": context
        })

        return approval

    async def _default_callback(self, request: dict) -> dict:
        print(f"\n⚠️  CHECKPOINT: {request['checkpoint']}")
        print(f"   {request['description']}")
        print(f"   Context: {json.dumps(request['context'], indent=2)[:500]}")

        response = input("Approve? (y/n/modify): ").strip().lower()
        if response == "y":
            return {"approved": True}
        elif response == "modify":
            modification = input("Enter modification: ")
            return {"approved": True, "modification": modification}
        return {"approved": False, "reason": "Human rejected"}

# Usage
checkpoint = HumanCheckpoint()
checkpoint.register(
    "high_cost_action",
    condition=lambda ctx: ctx.get("estimated_cost", 0) > 10.0,
    description="Agent wants to perform a high-cost LLM operation"
)
checkpoint.register(
    "external_api_call",
    condition=lambda ctx: ctx.get("action_type") == "external_api",
    description="Agent wants to call an external API"
)
checkpoint.register(
    "data_modification",
    condition=lambda ctx: ctx.get("action_type") in ("insert", "update", "delete"),
    description="Agent wants to modify data"
)
```

### Agent Sandboxing

```python
class SandboxedAgent:
    """Run agent with restricted permissions."""

    def __init__(self, agent, permissions: dict):
        self.agent = agent
        self.permissions = permissions
        self.audit_log = []

    def run(self, query: str) -> dict:
        # Wrap tool functions with permission checks
        original_tools = self.agent.tool_functions
        self.agent.tool_functions = self._wrap_tools(original_tools)

        try:
            result = self.agent.run(query)
            result["audit_log"] = self.audit_log
            return result
        finally:
            self.agent.tool_functions = original_tools

    def _wrap_tools(self, tools: dict) -> dict:
        wrapped = {}
        for name, func in tools.items():
            wrapped[name] = self._create_wrapper(name, func)
        return wrapped

    def _create_wrapper(self, name: str, func: callable) -> callable:
        def wrapper(**kwargs):
            # Check permissions
            if name not in self.permissions.get("allowed_tools", []):
                self.audit_log.append({
                    "action": "blocked",
                    "tool": name,
                    "reason": "Tool not in allowed list"
                })
                return f"Permission denied: {name} is not allowed"

            # Check rate limits
            recent_calls = sum(1 for log in self.audit_log
                             if log.get("tool") == name and log["action"] == "executed")
            max_calls = self.permissions.get("rate_limits", {}).get(name, 100)
            if recent_calls >= max_calls:
                return f"Rate limit exceeded for {name}"

            # Log and execute
            self.audit_log.append({
                "action": "executed",
                "tool": name,
                "args": {k: str(v)[:100] for k, v in kwargs.items()},
                "timestamp": datetime.utcnow().isoformat()
            })
            return func(**kwargs)

        return wrapper
```

---

## Agent Observability

### Tracing and Logging

```python
import uuid
from contextlib import contextmanager

class AgentTracer:
    """Trace agent execution for debugging and monitoring."""

    def __init__(self):
        self.traces = {}
        self.current_trace_id = None

    @contextmanager
    def trace(self, name: str, metadata: dict = None):
        """Start a new trace span."""
        trace_id = str(uuid.uuid4())
        parent_id = self.current_trace_id

        span = {
            "id": trace_id,
            "parent_id": parent_id,
            "name": name,
            "metadata": metadata or {},
            "start_time": datetime.utcnow().isoformat(),
            "children": [],
            "events": []
        }

        if parent_id and parent_id in self.traces:
            self.traces[parent_id]["children"].append(trace_id)

        self.traces[trace_id] = span
        self.current_trace_id = trace_id

        try:
            yield span
        except Exception as e:
            span["error"] = str(e)
            span["error_type"] = type(e).__name__
            raise
        finally:
            span["end_time"] = datetime.utcnow().isoformat()
            self.current_trace_id = parent_id

    def event(self, name: str, data: dict = None):
        """Record an event in the current trace."""
        if self.current_trace_id and self.current_trace_id in self.traces:
            self.traces[self.current_trace_id]["events"].append({
                "name": name,
                "data": data or {},
                "timestamp": datetime.utcnow().isoformat()
            })

    def get_trace_tree(self, trace_id: str) -> dict:
        """Get a trace and all its children as a tree."""
        trace = self.traces.get(trace_id)
        if not trace:
            return None

        result = dict(trace)
        result["children"] = [
            self.get_trace_tree(child_id)
            for child_id in trace["children"]
        ]
        return result

# Usage with agent
tracer = AgentTracer()

class TracedAgent:
    def __init__(self, agent, tracer: AgentTracer):
        self.agent = agent
        self.tracer = tracer

    def run(self, query: str) -> dict:
        with self.tracer.trace("agent_run", {"query": query}) as span:
            self.tracer.event("query_received", {"length": len(query)})

            result = self.agent.run(query)

            self.tracer.event("completed", {
                "steps": result.get("steps"),
                "answer_length": len(result.get("answer", ""))
            })

            return result
```

### Cost Tracking

```python
class CostTracker:
    """Track LLM API costs per agent execution."""

    PRICING = {
        "gpt-4o": {"input": 2.50 / 1_000_000, "output": 10.00 / 1_000_000},
        "gpt-4o-mini": {"input": 0.15 / 1_000_000, "output": 0.60 / 1_000_000},
        "claude-3-5-sonnet-20241022": {"input": 3.00 / 1_000_000, "output": 15.00 / 1_000_000},
        "text-embedding-3-small": {"input": 0.02 / 1_000_000, "output": 0},
    }

    def __init__(self):
        self.calls = []
        self.total_cost = 0

    def record(self, model: str, input_tokens: int, output_tokens: int, metadata: dict = None):
        pricing = self.PRICING.get(model, {"input": 0, "output": 0})
        cost = (input_tokens * pricing["input"]) + (output_tokens * pricing["output"])

        self.calls.append({
            "model": model,
            "input_tokens": input_tokens,
            "output_tokens": output_tokens,
            "cost": cost,
            "timestamp": datetime.utcnow().isoformat(),
            "metadata": metadata or {}
        })
        self.total_cost += cost
        return cost

    def summary(self) -> dict:
        by_model = {}
        for call in self.calls:
            model = call["model"]
            if model not in by_model:
                by_model[model] = {"calls": 0, "input_tokens": 0, "output_tokens": 0, "cost": 0}
            by_model[model]["calls"] += 1
            by_model[model]["input_tokens"] += call["input_tokens"]
            by_model[model]["output_tokens"] += call["output_tokens"]
            by_model[model]["cost"] += call["cost"]

        return {
            "total_cost": self.total_cost,
            "total_calls": len(self.calls),
            "by_model": by_model
        }
```

---

## Production Agent Patterns

### Agent State Machine

```python
from enum import Enum

class AgentState(Enum):
    IDLE = "idle"
    THINKING = "thinking"
    ACTING = "acting"
    WAITING_APPROVAL = "waiting_approval"
    ERROR = "error"
    COMPLETED = "completed"

class StatefulAgent:
    """Agent with explicit state management for production use."""

    def __init__(self, client: OpenAI, tools: list, tool_functions: dict):
        self.client = client
        self.tools = tools
        self.tool_functions = tool_functions
        self.state = AgentState.IDLE
        self.state_history = []
        self.context = {}

    def _transition(self, new_state: AgentState, reason: str = ""):
        self.state_history.append({
            "from": self.state.value,
            "to": new_state.value,
            "reason": reason,
            "timestamp": datetime.utcnow().isoformat()
        })
        self.state = new_state

    async def run(self, query: str) -> dict:
        self._transition(AgentState.THINKING, "Received query")
        self.context["query"] = query
        messages = [{"role": "user", "content": query}]

        for step in range(10):
            try:
                self._transition(AgentState.THINKING, f"Step {step + 1}")

                response = self.client.chat.completions.create(
                    model="gpt-4o",
                    messages=messages,
                    tools=self.tools,
                    temperature=0.1
                )

                choice = response.choices[0]

                if not choice.message.tool_calls:
                    self._transition(AgentState.COMPLETED, "Final answer generated")
                    return {
                        "answer": choice.message.content,
                        "state_history": self.state_history
                    }

                self._transition(AgentState.ACTING, f"Executing {len(choice.message.tool_calls)} tool(s)")
                messages.append(choice.message)

                for tool_call in choice.message.tool_calls:
                    result = self._execute_tool(tool_call)
                    messages.append({
                        "role": "tool",
                        "tool_call_id": tool_call.id,
                        "content": result
                    })

            except Exception as e:
                self._transition(AgentState.ERROR, str(e))
                return {"error": str(e), "state_history": self.state_history}

        self._transition(AgentState.COMPLETED, "Max steps reached")
        return {"answer": "Max steps reached", "state_history": self.state_history}

    def _execute_tool(self, tool_call) -> str:
        name = tool_call.function.name
        args = json.loads(tool_call.function.arguments)
        try:
            return str(self.tool_functions[name](**args))
        except Exception as e:
            return f"Error: {str(e)}"
```

### Persistent Agent with Checkpointing

```python
class PersistentAgent:
    """Agent that can be paused, saved, and resumed."""

    def __init__(self, agent_id: str, storage_backend):
        self.agent_id = agent_id
        self.storage = storage_backend
        self.messages = []
        self.step = 0
        self.metadata = {}

    def save_checkpoint(self):
        """Save current agent state."""
        self.storage.save(self.agent_id, {
            "messages": self.messages,
            "step": self.step,
            "metadata": self.metadata,
            "timestamp": datetime.utcnow().isoformat()
        })

    def load_checkpoint(self) -> bool:
        """Load agent state from storage."""
        data = self.storage.load(self.agent_id)
        if data:
            self.messages = data["messages"]
            self.step = data["step"]
            self.metadata = data["metadata"]
            return True
        return False

    async def run_with_checkpointing(self, query: str = None, checkpoint_every: int = 3):
        """Run agent with periodic checkpointing."""
        if query:
            self.messages.append({"role": "user", "content": query})

        while self.step < 20:  # Max steps
            self.step += 1

            # Periodic checkpoint
            if self.step % checkpoint_every == 0:
                self.save_checkpoint()

            # Execute step...
            # (agent logic here)

            break  # Placeholder

        self.save_checkpoint()
```

---

## Design Principles for Building Agents

1. **Start with a single tool, single loop.** ReAct is enough for 80% of use cases. Add complexity only when simple agents fail.

2. **Make tools atomic and idempotent.** Each tool should do one thing. Calling it twice with the same inputs should be safe.

3. **Always have a maximum step limit.** Agents can loop forever. Set a hard cap (5-10 steps for most tasks, 20 for complex research).

4. **Log everything.** Every thought, action, and observation. You cannot debug an agent you cannot observe.

5. **Build escape hatches.** Human-in-the-loop checkpoints for high-risk actions. Circuit breakers for runaway costs.

6. **Test with adversarial inputs.** Users will ask things your agent wasn't designed for. Plan for graceful degradation.

7. **Memory is expensive.** Vector memory adds latency and cost. Start with buffer memory and add complexity only when needed.

8. **Multi-agent adds multi-complexity.** A single well-designed agent often beats a poorly coordinated team. Justify the complexity.

9. **Validate tool outputs before passing them back.** Tool errors should be caught, not passed raw to the LLM.

10. **Design for observability from day one.** Tracing, cost tracking, and quality metrics are not afterthoughts — they're core features.
