// Package inject resolves secrets from a Provider and injects their values
// into structured targets or text templates.
//
// # Basic injection
//
// An Injector fetches a list of keys from a Provider and writes the resolved
// values to any Target implementation:
//
//	inj, err := inject.New(myProvider)
//	m := inject.MapTarget{}
//	errs := inj.Inject(ctx, []string{"db/pass", "api/key"}, m)
//
// # Template injection
//
// A TemplateInjector replaces {{ secret "key" }} placeholders embedded in
// arbitrary text (e.g. config files, scripts) with the live secret values:
//
//	ti, err := inject.NewTemplateInjector(myProvider)
//	out, err := ti.Render(ctx, `DSN=postgres://user:{{ secret "db/pass" }}@host/db`)
//
// # Prefix stripping
//
// Use WithPrefix to strip a common path prefix from keys before writing to
// the target, keeping destination keys clean:
//
//	inj, _ := inject.New(myProvider, inject.WithPrefix("prod/"))
package inject
