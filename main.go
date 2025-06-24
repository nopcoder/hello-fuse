package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type HelloRoot struct {
	fs.Inode
}

func (r *HelloRoot) OnAdd(ctx context.Context) {
	ch := r.NewPersistentInode(
		ctx, &fs.MemRegularFile{
			Data: []byte("file.txt"),
			Attr: fuse.Attr{
				Mode: 0644,
			},
		}, fs.StableAttr{Ino: 2})
	r.AddChild("file.txt", ch, false)
}

func (r *HelloRoot) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0755
	return 0
}

var (
	_ = (fs.NodeGetattrer)((*HelloRoot)(nil))
	_ = (fs.NodeOnAdder)((*HelloRoot)(nil))
)

func resolveUIDGID(uid int64, gid int64) (uint32, uint32, error) {
	currentUID, currentGID, err := getCurrentUIDGID()
	if err != nil {
		return 0, 0, err
	}
	if uid == -1 {
		uid = int64(currentUID)
	}
	if gid == -1 {
		gid = int64(currentGID)
	}
	return uint32(uid), uint32(gid), nil //nolint:gosec
}

func getCurrentUIDGID() (uid, gid uint32, err error) {
	var currentUser *user.User
	currentUser, err = user.Current()
	if err != nil {
		return
	}
	var uidInt, gidInt int
	uidInt, err = strconv.Atoi(currentUser.Uid)
	if err != nil {
		return
	}
	gidInt, err = strconv.Atoi(currentUser.Gid)
	if err != nil {
		return
	}
	return uint32(uidInt), uint32(gidInt), nil //nolint:gosec
}

func main() {
	debug := flag.Bool("debug", false, "print debug data")

	// fs.Options
	entryTimeout := flag.Duration("entryTimeout", time.Second, "fuse entry timeout")
	attrTimeout := flag.Duration("attrTimeout", time.Second, "fuse attribute timeout")
	negativeTimeout := flag.Duration("negativeTimeout", time.Second, "fuse negative entry timeout")
	firstAutomaticIno := flag.Uint64("firstAutomaticIno", 0, "first automatic inode number")
	nullPermissions := flag.Bool("nullPermissions", false, "support null permissions")
	uid := flag.Int64("uid", -1, "user id")
	gid := flag.Int64("gid", -1, "group id")
	// fuse.MountOptions
	allowOther := flag.Bool("allowOther", false, "allow other users to access the file system")
	maxBackground := flag.Int("maxBackground", 12, "max number of background requests")
	maxWrite := flag.Int("maxWrite", 0, "max size for write requests")
	maxReadAhead := flag.Int("maxReadAhead", 0, "max read ahead size")
	ignoreSecurityLabels := flag.Bool("ignoreSecurityLabels", false, "ignore security labels")
	rememberInodes := flag.Bool("rememberInodes", false, "remember inodes")
	fsName := flag.String("fsName", "", "filesystem name")
	name := flag.String("name", "", "mount name")
	singleThreaded := flag.Bool("singleThreaded", false, "single threaded")
	disableXAttrs := flag.Bool("disableXAttrs", false, "disable extended attributes")
	enableLocks := flag.Bool("enableLocks", false, "enable file locks")
	enableSymlinkCaching := flag.Bool("enableSymlinkCaching", false, "enable symlink caching")
	explicitDataCacheControl := flag.Bool("explicitDataCacheControl", false, "explicit data cache control")
	syncRead := flag.Bool("syncRead", false, "synchronous read")
	directMount := flag.Bool("directMount", false, "direct mount")
	directMountStrict := flag.Bool("directMountStrict", false, "strict direct mount")
	directMountFlags := flag.Uint("directMountFlags", 0, "direct mount flags")
	enableAcl := flag.Bool("enableAcl", false, "enable ACL support")
	disableReadDirPlus := flag.Bool("disableReadDirPlus", false, "disable readdirplus")
	disableSplice := flag.Bool("disableSplice", false, "disable splice")
	maxStackDepth := flag.Int("maxStackDepth", 1, "maximum stacking depth")
	idMappedMount := flag.Bool("idMappedMount", false, "ID-mapped mount")
	optionsStr := flag.String("options", "", "comma-separated mount options")
	mountTimeout := flag.Duration("mountTimeout", 5*time.Second, "timeout for mounting the filesystem")

	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Printf("Usage:\n  hello-fuse [flags] MOUNTPOINT\n")
		return
	}
	ruid, rgid, err := resolveUIDGID(*uid, *gid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving UID/GID: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Using UID: '%d', GID: '%d'\n", ruid, rgid)
	var options []string
	if *optionsStr != "" {
		options = strings.Split(*optionsStr, ",")
	}

	opts := &fs.Options{
		Logger:            log.New(os.Stdout, "", log.LstdFlags),
		EntryTimeout:      entryTimeout,
		AttrTimeout:       attrTimeout,
		NegativeTimeout:   negativeTimeout,
		FirstAutomaticIno: *firstAutomaticIno,
		NullPermissions:   *nullPermissions,
		UID:               ruid,
		GID:               rgid,
		MountOptions: fuse.MountOptions{
			Debug:                    *debug,
			AllowOther:               *allowOther,
			Options:                  options,
			MaxBackground:            *maxBackground,
			MaxWrite:                 *maxWrite,
			MaxReadAhead:             *maxReadAhead,
			IgnoreSecurityLabels:     *ignoreSecurityLabels,
			RememberInodes:           *rememberInodes,
			FsName:                   *fsName,
			Name:                     *name,
			SingleThreaded:           *singleThreaded,
			DisableXAttrs:            *disableXAttrs,
			EnableLocks:              *enableLocks,
			EnableSymlinkCaching:     *enableSymlinkCaching,
			ExplicitDataCacheControl: *explicitDataCacheControl,
			SyncRead:                 *syncRead,
			DirectMount:              *directMount,
			DirectMountStrict:        *directMountStrict,
			DirectMountFlags:         uintptr(*directMountFlags),
			EnableAcl:                *enableAcl,
			DisableReadDirPlus:       *disableReadDirPlus,
			DisableSplice:            *disableSplice,
			MaxStackDepth:            *maxStackDepth,
			IDMappedMount:            *idMappedMount,
		},
	}

	var (
		server   *fuse.Server
		mountErr error
	)
	done := make(chan struct{})
	// Signal handling for graceful shutdown to call umount
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	mountpoint := flag.Arg(0)
	go func() {
		server, mountErr = fs.Mount(mountpoint, &HelloRoot{}, opts)
		close(done)
	}()
	select {
	case <-done:
		if mountErr != nil {
			fmt.Fprintf(os.Stderr, "Mount fail: %v\n", mountErr)
			os.Exit(1)
		}
	case <-time.After(*mountTimeout):
		fmt.Fprintf(os.Stderr, "ERROR: Mount failed timed out after %v\nHint: Perhaps mount directory busy? try runnning 'umount %s'\n", *mountTimeout, mountpoint)
		os.Exit(1)
	}
	// wait group for server
	wg := &sync.WaitGroup{}
	wg.Add(1)
	// Handle Ctrl+C or shell close
	go func() {
		sig := <-sigCh
		fmt.Printf("Received signal %v, Closing gracefully\n", sig)
		cmd := exec.Command("umount", mountpoint)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to unmount: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	go func() {
		server.Wait()
		wg.Done()
	}()

	// verify mount by trying to stat a file
	go tryStatFile(mountpoint)
	fmt.Println("Mount ready")
	wg.Wait()
}

func tryStatFile(mountpoint string) {
	var err error
	for range 3 { // try 3 times
		_, err = os.Stat(mountpoint + "/file.txt")
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Mount failed, error stating file: %v\n", err)
		os.Exit(1)
	}
}
