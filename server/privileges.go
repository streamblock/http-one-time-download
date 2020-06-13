package main

import (
	"errors"
	"log"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

/*
#include <sys/types.h>
#include <stdint.h>
#include <unistd.h>

static int switch_uid_gid(uint32_t uid, uint32_t gid)
{
	if (setgid((gid_t)gid) < 0) {
        return -1;
	}
	if (setuid((uid_t)uid) < 0) {
		return -1;
	}
	return 0;
}
*/
import "C"

func chowmPath(path string, strUID string) error {
	if syscall.Getuid() == 0 {
		userInfo, err := user.Lookup(strUID)
		if err != nil {
			return errors.New("user not found or other error: " + err.Error())
		}
		uid, err := strconv.ParseInt(userInfo.Uid, 10, 32)
		if err != nil {
			return errors.New("cannot convert UID to string: " + err.Error())
		}
		gid, err := strconv.ParseInt(userInfo.Gid, 10, 32)
		if err != nil {
			return errors.New("cannot convert GID to string: " + err.Error())
		}

		err = os.Chown(path, int(uid), int(gid))
		if err != nil {
			return errors.New("chowm failed: " + err.Error())
		}
		return nil
	}

	return errors.New("not running as root, cannot chown")
}

func dropPrivileges(strUID string) error {
	if syscall.Getuid() == 0 {
		log.Printf("Running as root, downgrading to user %v\n", strUID)

		userInfo, err := user.Lookup(strUID)
		if err != nil {
			return errors.New("user not found or other error: " + err.Error())
		}
		uid, err := strconv.ParseInt(userInfo.Uid, 10, 32)
		if err != nil {
			return errors.New("cannot convert UID to string: " + err.Error())
		}
		gid, err := strconv.ParseInt(userInfo.Gid, 10, 32)
		if err != nil {
			return errors.New("cannot convert GID to string: " + err.Error())
		}

		ret := C.switch_uid_gid(C.uint32_t(uid), C.uint32_t(gid))
		if ret < 0 {
			return errors.New("unable to set new UIG/GID")
		}

		log.Printf("Privileges drooped to %v (%v)\n", strUID, syscall.Getuid())
		return nil
	}
	return errors.New("not running as root, cannot change user")
}