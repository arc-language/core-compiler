

1: structs & classes 
  - check the lang readme
  - inine and non inline support

3: importing and modules/packages, core modules
  - core modules use pure syscall, for macos they use linker 
  - need to add let result = syscall(100, ...) support to the parser, codegen and builder
  - and then your io modules can just wrap around with types support
  - and you also need module for platform checker
  - 	"github.com/arc-language/io"
  - io.printf()

4: linker
 - 	"github.com/arc-language/core-linker"
 -  mainly replace gcc and read the .so files and patch offsets directly instead 
