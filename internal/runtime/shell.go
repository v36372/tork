package runtime

import (
	"context"

	"fmt"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/runabol/tork"
	"github.com/runabol/tork/internal/syncx"
)

type ShellRuntime struct {
	cmds *syncx.Map[string, *exec.Cmd]
}

func NewShellRuntime() *ShellRuntime {
	return &ShellRuntime{
		cmds: new(syncx.Map[string, *exec.Cmd]),
	}
}

func (r *ShellRuntime) Run(ctx context.Context, t *tork.Task) error {
	if t.ID == "" {
		return errors.New("task id is required")
	}
	if len(t.Mounts) > 0 {
		return errors.New("mounts are not supported on shell runtime")
	}
	if len(t.Entrypoint) > 0 {
		return errors.New("entrypoint is not supported on shell runtime")
	}
	if t.Image != "" {
		return errors.New("image is not supported on shell runtime")
	}
	if t.Limits != nil && (t.Limits.CPUs != "" || t.Limits.Memory != "") {
		return errors.New("limits are not supported on shell runtime")
	}
	if len(t.Networks) > 0 {
		return errors.New("networks are not supported on shell runtime")
	}
	if t.Registry != nil {
		return errors.New("registry is not supported on shell runtime")
	}
	defer r.cmds.Delete(t.ID)

	workdir, err := os.MkdirTemp("", "tork")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workdir)

	if err := os.WriteFile(fmt.Sprintf("%s/stdout", workdir), []byte{}, os.ModePerm); err != nil {
		return errors.Wrapf(err, "error writing the entrypoint")
	}

	for filename, contents := range t.Files {
		filename = fmt.Sprintf("%s/%s", workdir, filename)
		if err := os.WriteFile(filename, []byte(contents), os.ModePerm); err != nil {
			return errors.Wrapf(err, "error writing file: %s", filename)
		}
	}

	env := []string{}
	for name, value := range t.Env {
		env = append(env, fmt.Sprintf("%s=%s", name, value))
	}
	env = append(env, fmt.Sprintf("TORK_OUTPUT=%s/stdout", workdir))

	var cmd *exec.Cmd
	if len(t.CMD) > 0 {
		cmd = exec.CommandContext(ctx, t.CMD[0], t.CMD[1:]...)
	} else {
		if err := os.WriteFile(fmt.Sprintf("%s/entrypoint", workdir), []byte(t.Run), os.ModePerm); err != nil {
			return errors.Wrapf(err, "error writing the entrypoint")
		}
		cmd = exec.CommandContext(ctx, "bash", "-c", "entrypoint")
	}
	cmd.Dir = workdir
	cmd.Env = env

	if err := cmd.Start(); err != nil {
		return err
	}

	r.cmds.Set(t.ID, cmd)

	if err := cmd.Wait(); err != nil {
		return err
	}

	output, err := os.ReadFile(fmt.Sprintf("%s/stdout", workdir))
	if err != nil {
		return errors.Wrapf(err, "error reading the task output")
	}

	t.Result = string(output)

	return nil
}

func (r *ShellRuntime) Stop(ctx context.Context, t *tork.Task) error {
	proc, ok := r.cmds.Get(t.ID)
	if !ok {
		return nil
	}
	if err := proc.Cancel(); err != nil {
		return errors.Wrapf(err, "error stopping process for task: %s", t.ID)
	}
	return nil
}

func (r *ShellRuntime) HealthCheck(ctx context.Context) error {
	return nil
}
