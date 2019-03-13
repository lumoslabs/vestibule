package main

import (
	"os"
	"syscall"

	"github.com/opencontainers/runc/libcontainer/user"
)

// this function comes from libcontainer/init_linux.go
// we don't use that directly because we don't want the whole namespaces package imported here
// (also, because we need minor modifications and it's not even exported)

// SetupUser changes the groups, gid, and uid for the user inside the container
func SetupUser(usr *user.ExecUser) error {
	if err := syscall.Setgroups(usr.Sgids); err != nil {
		return err
	}
	if err := syscall.Setgid(usr.Gid); err != nil {
		return err
	}
	if err := syscall.Setuid(usr.Uid); err != nil {
		return err
	}
	// if we didn't get HOME already, set it based on the user's HOME
	if home := os.Getenv("HOME"); home == "" {
		if err := os.Setenv("HOME", usr.Home); err != nil {
			return err
		}
	}
	return nil
}
