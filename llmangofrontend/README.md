# LLMango Frontend

## Template System Migration

This project is in the process of migrating from Go's standard `html/template` to the [templ](https://templ.guide) templating library. The migration offers several benefits:

- **Type Safety**: Catch errors at compile time rather than runtime
- **IDE Support**: Better autocompletion and refactoring tools
- **Modularity**: More reusable and maintainable components 
- **Developer Experience**: More Go-like syntax and patterns

## Current Status

- The existing Go template-based routes are fully functional and remain the default
- New templ-based routes are available under the `/templ` prefix
- Migration is being done incrementally, component by component

## Using templ Templates

To generate Go code from the `.templ` files, run:

```bash
templ generate
```

This will process all `.templ` files in the project and generate the corresponding Go code.

## Directory Structure

- `templates/`: Original Go templates (string constants)
- `templates_templ/`: New templ-based templates (`.templ` files)

## Developing with templ

1. Create or modify `.templ` files in the `templates_templ/` directory
2. Run `templ generate` to generate Go code
3. Access the templ-based version of a page at `/templ/[page]` 

## Testing

Both template systems are running in parallel to allow for easy comparison and testing. Visit:

- Standard route: `/home`
- Templ-based route: `/templ/home`

## Migration Plan

See the detailed migration plan in `templRefactor.md` for information about the conversion process and timeline. 