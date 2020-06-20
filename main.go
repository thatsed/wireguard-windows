/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019 WireGuard LLC. All Rights Reserved.
 */

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/windows"

	"golang.zx2c4.com/wireguard/windows/elevate"
	"golang.zx2c4.com/wireguard/windows/l18n"
	"golang.zx2c4.com/wireguard/windows/manager"
	"golang.zx2c4.com/wireguard/windows/ringlogger"
	"golang.zx2c4.com/wireguard/windows/tunnel"
	"golang.zx2c4.com/wireguard/windows/updater"
)

func fatal(v ...interface{}) {
	windows.MessageBox(0, windows.StringToUTF16Ptr(fmt.Sprint(v...)), windows.StringToUTF16Ptr(l18n.Sprintf("Error")), windows.MB_ICONERROR)
	os.Exit(1)
}

func fatalf(format string, v ...interface{}) {
	fatal(l18n.Sprintf(format, v...))
}

func info(title string, format string, v ...interface{}) {
	windows.MessageBox(0, windows.StringToUTF16Ptr(l18n.Sprintf(format, v...)), windows.StringToUTF16Ptr(title), windows.MB_ICONINFORMATION)
}

func usage() {
	var flags = [...]string{
		l18n.Sprintf("(no argument): elevate and install manager service"),
		"/installmanagerservice",
		"/installtunnelservice CONFIG_PATH",
		"/uninstallmanagerservice",
		"/uninstalltunnelservice TUNNEL_NAME",
		"/managerservice",
		"/tunnelservice CONFIG_PATH",
		"/dumplog OUTPUT_PATH",
		"/update [LOG_FILE]",
	}
	builder := strings.Builder{}
	for _, flag := range flags {
		builder.WriteString(fmt.Sprintf("    %s\n", flag))
	}
	info(l18n.Sprintf("Command Line Options"), "Usage: %s [\n%s]", os.Args[0], builder.String())
	os.Exit(1)
}

func checkForWow64() {
	var b bool
	err := windows.IsWow64Process(windows.CurrentProcess(), &b)
	if err != nil {
		fatalf("Unable to determine whether the process is running under WOW64: %v", err)
	}
	if b {
		fatalf("You must use the 64-bit version of WireGuard on this computer.")
	}
}

func execElevatedManagerServiceInstaller() error {
	path, err := os.Executable()
	if err != nil {
		return err
	}
	err = elevate.ShellExecute(path, "/installmanagerservice", "", windows.SW_SHOW)
	if err != nil {
		return err
	}
	os.Exit(0)
	return windows.ERROR_ACCESS_DENIED // Not reached
}

func main() {
	checkForWow64()

	switch os.Args[1] {
	case "/installmanagerservice":
		if len(os.Args) != 2 {
			usage()
		}
		err := manager.InstallManager()
		if err != nil {
			fatal(err)
		}
	case "/uninstallmanagerservice":
		if len(os.Args) != 2 {
			usage()
		}
		err := manager.UninstallManager()
		if err != nil {
			fatal(err)
		}
		return
	case "/installtunnelservice":
		if len(os.Args) != 3 {
			usage()
		}
		err := manager.InstallTunnel(os.Args[2])
		if err != nil {
			fatal(err)
		}
		return
	case "/uninstalltunnelservice":
		if len(os.Args) != 3 {
			usage()
		}
		err := manager.UninstallTunnel(os.Args[2])
		if err != nil {
			fatal(err)
		}
		return
	case "/dumplog":
		if len(os.Args) != 3 {
			usage()
		}
		file, err := os.Create(os.Args[2])
		if err != nil {
			fatal(err)
		}
		defer file.Close()
		err = ringlogger.DumpTo(file, true)
		if err != nil {
			fatal(err)
		}
		return
	}
	usage()
}
