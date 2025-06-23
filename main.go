package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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

func main() {
	debug := flag.Bool("debug", false, "print debug data")

	// fs.Options
	entryTimeout := flag.Duration("entryTimeout", time.Second, "fuse entry timeout")
	attrTimeout := flag.Duration("attrTimeout", time.Second, "fuse attribute timeout")
	negativeTimeout := flag.Duration("negativeTimeout", time.Second, "fuse negative entry timeout")
	firstAutomaticIno := flag.Uint64("firstAutomaticIno", 0, "first automatic inode number")
	nullPermissions := flag.Bool("nullPermissions", false, "support null permissions")
	uid := flag.Uint("uid", 0, "user id")
	gid := flag.Uint("gid", 0, "group id")

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

	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Printf("Usage:\n  hello-fuse MOUNTPOINT\n")
		return
	}

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
		UID:               uint32(*uid),
		GID:               uint32(*gid),
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
	server, err := fs.Mount(flag.Arg(0), &HelloRoot{}, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Mount fail: %v\n", err)
		os.Exit(1)
	}
	server.Wait()
}
