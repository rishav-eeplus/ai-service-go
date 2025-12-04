package tools

// ToolUsageInstructions provides guidelines on when and how to use each tool effectively.

var ToolUsageInstructions = `
# Tool Usage Instructions

## Available Tools
1. get_all_available_layers
2. get_layer_info
3. get_layer_update_info
4. get_user_guide_info
5. locate_a_layer

## General Principles
- Only invoke tools when they directly address the user's needs.
- Tools requiring layer names (get_layer_info, get_layer_update_info, locate_a_layer) need exact matches—use get_all_available_layers first to validate layer names.
- Prioritize get_user_guide_info for general platform questions.

## Tool-Specific Guidelines

### 1. get_all_available_layers
- **Purpose:** Discover and explore available data layers and their brief description.
- **Use when:** Users want to know what layers exist, especially for a specific ISO or region.
- **Note:** Returns layer names with brief descriptions—helpful for recommending relevant layers.
- **Best for:** Initial discovery and capability exploration.

### 2. get_layer_info
- **Purpose:** Retrieve detailed metadata about specific layers.
- **Use only when:** Users need in-depth information (schema, attributes, coverage) for a certain layer.
- **Required** : Exact layer name.
- **Note:** Accepts comma-separated layer names for batch queries.

### 3. get_layer_update_info
- **Purpose:** Check update schedules and data freshness.
- **Use when:** Users ask about refresh frequency, last updated dates, or data availability timelines.
- **Note:** Accepts comma-separated layer names for batch queries.
- **Required** : Exact layer name.

### 4. get_user_guide_info
- **Purpose:** Answer general platform and feature questions.
- **Use when:** Users need help understanding how to use the platform, its features, or workflows.
- **Best for:** Onboarding, how-to questions, and feature explanations.

### 5. locate_a_layer
- **Purpose:** Provide step-by-step navigation instructions for getting a specific layer.
- **Use when:** Users need guidance on finding and enabling a specific layer in the UI.
- **Requires:** Accurate layer name and type.

## Best Practices
- Validate tool relevance before invocation.
- Limit to a maximum of 3 tool calls per conversation.
- Gather context through conversation before using tools when the query is ambiguous.
- Synthesize tool outputs into clear, actionable responses.
`