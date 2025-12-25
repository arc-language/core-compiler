

1: structs & classes 
  - check the lang readme
  - inine and non inline support

2: importing and modules/packages, core modules
  - core modules use pure syscall, for macos they use linker 
  - need to add let result = syscall(100, ...) support to the parser, codegen and builder
  - and then your io modules can just wrap around with types support
  - and you also need module for platform checker
  - 	"github.com/arc-language/io"
  - io.printf()

2:
Rule of Thumb
Does it hold an OS resource that needs cleanup?

YES → class (use deinit to cleanup)
NO → struct (just data)

3: linker
 - 	"github.com/arc-language/core-linker"
 -  mainly replace gcc and read the .so files and patch offsets directly instead


4: more core modules 

import "github.com/arc-language/ai"

// Unified Arc interface
let model = ai.Model{
    provider: "openai",
    model: "gpt-4o",
    api_key: env.get("OPENAI_KEY")
}

let model = ai.Model{
    provider: "xai", 
    model: "grok-2",
    api_key: env.get("XAI_KEY")
}

let model = ai.Model{
    provider: "anthropic",
    model: "claude-sonnet-4.5",
    api_key: env.get("ANTHROPIC_KEY")
}

let model = ai.Model{
    provider: "huggingface",
    model: "meta-llama/Llama-3-8B",
    local: true  // Self-hosted
}

// Same interface for all
let response = model.generate("Hello, world!")

- ai models
- ui
- http 
