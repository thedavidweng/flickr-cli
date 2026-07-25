package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/config"
	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/thedavidweng/flickr-cli/internal/safety"
)

func newRenderer(app *AppContext, cmd *cobra.Command) output.Renderer {
	return output.Renderer{
		Out:     cmd.OutOrStdout(),
		Err:     cmd.ErrOrStderr(),
		JSON:    app.JSON,
		Pretty:  app.Pretty,
		Compact: app.Compact,
		Full:    app.Full,
		Quiet:   app.Quiet,
		NoColor: app.NoColor,
		Verbose: app.Verbose,
	}
}

// CmdContext bundles everything a command handler needs.
type CmdContext struct {
	App    *AppContext
	Cmd    *cobra.Command
	Args   []string
	Client *flickr.Client
	Config *config.Config
	R      output.Renderer
	Meta   output.RuntimeMetaInput
}

type CmdFunc func(ctx *CmdContext) error

// withAuth wraps a CmdFunc: loads config, creates client, checks auth.
func withAuth(command string, fn CmdFunc) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   command,
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		client, cfg, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}
		return fn(&CmdContext{App: app, Cmd: cmd, Args: args, Client: client, Config: cfg, R: r, Meta: meta})
	}
}

// mutationSpec parameterizes runMutation: safety classification for the gate
// plus the --dry-run rendering for one remote mutation.
type mutationSpec struct {
	Command  string
	Method   string
	Resource map[string]any
	PlanMsg  string
	PlanData map[string]any
	PlanFunc func() (map[string]any, error)
}

// runMutation centralizes the safety gate, dry-run/planned rendering, and
// audit logging shared by every remote mutation. run performs the API call(s),
// renders its own success/failure human output, and returns the success
// payload (or a rendered error). A committed write is always recorded to the
// audit log; a failed audit write is fatal.
func (ctx *CmdContext) runMutation(spec mutationSpec, run func() (any, error)) error {
	mutation := safety.Mutation{
		Command:  spec.Command,
		Method:   spec.Method,
		Risk:     safety.ClassifyRisk(spec.Command),
		Resource: spec.Resource,
	}
	gate := safety.Check(safety.GateInput{
		ReadOnly: ctx.App.ReadOnly,
		DryRun:   ctx.App.DryRun,
		Confirm:  ctx.App.Confirm,
	}, mutation)
	if gate.Error != nil {
		return ctx.R.Failure(ctx.Meta, *gate.Error)
	}
	if gate.Planned {
		if spec.PlanMsg != "" {
			ctx.R.Human("%s", spec.PlanMsg)
		}
		data := map[string]any{"planned": true}
		for k, v := range spec.PlanData {
			data[k] = v
		}
		if spec.PlanFunc != nil {
			extra, err := spec.PlanFunc()
			if err != nil {
				return err
			}
			for k, v := range extra {
				data[k] = v
			}
		}
		return ctx.R.Success(ctx.Meta, data, nil)
	}

	data, runErr := run()
	if runErr != nil {
		_ = ctx.appendAudit(mutation, "error", runErr)
		return runErr
	}
	if err := ctx.appendAudit(mutation, "success", nil); err != nil {
		return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFilesystem, "audit write failed: %v", err))
	}
	return ctx.R.Success(ctx.Meta, data, nil)
}

func (ctx *CmdContext) appendAudit(m safety.Mutation, result string, opErr error) error {
	ev := &safety.AuditEvent{
		RequestID: ctx.App.RequestID,
		Profile:   ctx.App.Profile,
		Command:   m.Command,
		Method:    m.Method,
		Resource:  m.Resource,
		Confirmed: ctx.App.Confirm,
		Result:    result,
	}
	if opErr != nil {
		ev.Error = opErr.Error()
	}
	return safety.Append(ctx.auditLogPath(), ev)
}

func (ctx *CmdContext) auditLogPath() string {
	if ctx.Config != nil {
		if p, err := ctx.Config.GetProfile(ctx.App.Profile); err == nil && p.AuditLogPath != "" {
			return p.AuditLogPath
		}
	}
	return config.DefaultAuditLogPath(ctx.App.Profile)
}

func getClient(app *AppContext) (*flickr.Client, *config.Config, error) {
	cfgPath := app.ConfigFile
	if cfgPath == "" {
		cfgPath = config.DefaultConfigPath()
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, nil, fmt.Errorf("not configured. Run 'flickr auth login' to get started")
	}

	profile, err := cfg.GetProfile(app.Profile)
	if err != nil {
		return nil, cfg, fmt.Errorf("not authenticated. Run 'flickr auth login' to get started")
	}

	creds, err := config.CredentialsFromProfileAndEnv(profile)
	if err != nil {
		return nil, cfg, err
	}
	client := flickr.NewClient(creds.APIKey, creds.APISecret, creds.OAuthToken, creds.OAuthTokenSecret)
	client.Retries = app.Retries
	client.RequestInterval = app.RequestInterval
	applyEndpointOverrides(client, &profile.Endpoints)

	return client, cfg, nil
}

// applyEndpointOverrides copies any non-empty endpoint override from the
// profile onto the client (used for testing and self-hosted instances).
func applyEndpointOverrides(client *flickr.Client, e *config.Endpoints) {
	for _, o := range []struct {
		src string
		dst *string
	}{
		{e.REST, &client.Endpoints.REST},
		{e.Upload, &client.Endpoints.Upload},
		{e.RequestToken, &client.Endpoints.RequestToken},
		{e.Authorize, &client.Endpoints.Authorize},
		{e.AccessToken, &client.Endpoints.AccessToken},
	} {
		if o.src != "" {
			*o.dst = o.src
		}
	}
}

func requireAuth(r *output.Renderer, meta output.RuntimeMetaInput, client *flickr.Client) error {
	if !client.IsAuthenticated() {
		return r.Failure(meta, output.ErrorWithDetails(
			model.ErrAuthRequired,
			"Authentication required. Run 'flickr auth login' to authenticate.",
			map[string]any{"profile": meta.Profile},
		))
	}
	return nil
}
