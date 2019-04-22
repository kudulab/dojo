// +build darwin

package main

import "syscall"

const ioctlReadTermios = syscall.TIOCGETA