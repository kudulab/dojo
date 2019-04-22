// +build linux

package main

import "syscall"

const ioctlReadTermios = syscall.TCGETS