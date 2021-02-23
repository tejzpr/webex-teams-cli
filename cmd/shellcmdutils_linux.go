package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"syscall"
	"time"
)

func (app *ShellCMDApplication) startUserInitiatedKillRoutine() {
	go func() {
		for terminateCMDID := range app.TerminateCommandIDChannel {
			var exeCmd executedCommand
			if tmp, ok := app.ExecutedCommands.Get(terminateCMDID); ok {
				exeCmd = tmp.(executedCommand)
				if exeCmd.cmd.ProcessState == nil {
					err := syscall.Kill(-exeCmd.cmd.Process.Pid, syscall.SIGTERM)
					if err != nil {
						app.sendError(err)
						app.stdHandler(exeCmd.cmdObj, "", err.Error())
					} else {
						app.stdHandler(exeCmd.cmdObj, fmt.Sprintf("Command [%s] was killed after recieving a kill request", exeCmd.cmdObj.CMD), "")
					}
					return
				}
			} else {
				app.stdHandler(exeCmd.cmdObj, fmt.Sprintf("Command ID %s is not valid", terminateCMDID), "")
			}

		}
	}()
}

func (app *ShellCMDApplication) shellout(cmdObj command) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command(app.ShellToUse, "-c", cmdObj.CMD)

	// Create a process group & create namespace
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Start()
	app.sendShellResponseMessage(app.sendToLog(fmt.Sprintf("Processing command (ID: %s): `%s`", cmdObj.ID, cmdObj.CMD), cmdObj.TRIGGEREDBYEMAIL), nil)
	app.ExecutedCommands.Set(cmdObj.ID, executedCommand{cmd, cmdObj})

	go func(cmd *exec.Cmd) {
		if app.ProcessTimeout != 0 {
			time.Sleep(app.ProcessTimeout)
			if cmd.ProcessState == nil {
				err := syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
				if err != nil {
					app.sendError(err)
					app.stdHandler(cmdObj, "", err.Error())
				} else {
					app.stdHandler(cmdObj, fmt.Sprintf("Process was killed after exceeding Timeout of %v", app.ProcessTimeout), "")
				}
				return
			}
		}
	}(cmd)

	err := cmd.Wait()
	if err != nil {
		app.sendError(err)
		app.stdHandler(cmdObj, "", err.Error())
	} else {
		if stdout.Len() > 0 || stderr.Len() > 0 {
			if stdout.Len() > 0 {
				if stdout.Len() <= 5000 {
					app.sendToLog(fmt.Sprintf("Response STDOUT sent for command (ID: %s): `%s`", cmdObj.ID, cmdObj.CMD), cmdObj.TRIGGEREDBYEMAIL)
					app.sendShellResponseMessage(fmt.Sprintf("STDOUT Response for command (ID: %s): \n``` \n %s``` ", cmdObj.ID, stdout.String()), nil)
				} else {
					stdoutfile, err := ioutil.TempFile(app.DownloadsDir, fmt.Sprintf("STDOUT-%s*.txt", cmdObj.ID))
					if err != nil {
						app.sendError(err)
					} else {
						stdoutfile.Write(stdout.Bytes())
						app.sendToLog(fmt.Sprintf("Response STDOUT sent for command (ID: %s): `%s`", cmdObj.ID, cmdObj.CMD), cmdObj.TRIGGEREDBYEMAIL)
						app.sendShellResponseMessage(fmt.Sprintf("STDOUT Response for command (ID: %s): `%s`", cmdObj.ID, cmdObj.CMD), stdoutfile)
					}
				}
			}

			if stderr.Len() > 0 {
				if stderr.Len() <= 5000 {
					app.sendToLog(fmt.Sprintf("Response STDERR sent for command (ID: %s): `%s`", cmdObj.ID, cmdObj.CMD), cmdObj.TRIGGEREDBYEMAIL)
					app.sendShellResponseMessage(fmt.Sprintf("STDERR Response for command (ID: %s): \n``` \n %s``` ", cmdObj.ID, stderr.String()), nil)
				} else {
					stderrfile, err := ioutil.TempFile(app.DownloadsDir, fmt.Sprintf("STDERR-%s*.txt", cmdObj.ID))
					if err != nil {
						app.sendError(err)
					} else {
						stderrfile.Write(stderr.Bytes())
						app.sendToLog(fmt.Sprintf("Response STDERR sent for command (ID: %s): `%s`", cmdObj.ID, cmdObj.CMD), cmdObj.TRIGGEREDBYEMAIL)
						app.sendShellResponseMessage(fmt.Sprintf("STDERR Response for command (ID: %s): `%s`", cmdObj.ID, cmdObj.CMD), stderrfile)
					}
				}
			}
		} else {
			app.sendToLog(fmt.Sprintf("Response sent for command (ID: %s): `%s`", cmdObj.ID, cmdObj.CMD), cmdObj.TRIGGEREDBYEMAIL)
			app.sendShellResponseMessage(fmt.Sprintf("Command (ID: %s): `%s` completed successfully", cmdObj.ID, cmdObj.CMD), nil)
		}
	}
}
