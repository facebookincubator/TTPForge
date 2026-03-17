/*
Copyright © 2025-present, Meta Platforms, Inc. and affiliates
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package backends

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"github.com/spf13/afero"
)

// SFTPFs implements afero.Fs over an SFTP connection.
type SFTPFs struct {
	client *sftp.Client
}

// NewSFTPFs creates a new SFTP-backed afero.Fs.
func NewSFTPFs(client *sftp.Client) afero.Fs {
	return &SFTPFs{client: client}
}

// Create creates a file on the remote host.
func (fs *SFTPFs) Create(name string) (afero.File, error) {
	f, err := fs.client.Create(name)
	if err != nil {
		return nil, err
	}
	return &SFTPFile{file: f, name: name, client: fs.client}, nil
}

// Mkdir creates a directory on the remote host.
func (fs *SFTPFs) Mkdir(name string, _ os.FileMode) error {
	return fs.client.Mkdir(name)
}

// MkdirAll creates a directory path on the remote host.
func (fs *SFTPFs) MkdirAll(path string, _ os.FileMode) error {
	return fs.client.MkdirAll(path)
}

// Open opens a file on the remote host for reading.
func (fs *SFTPFs) Open(name string) (afero.File, error) {
	f, err := fs.client.Open(name)
	if err != nil {
		return nil, err
	}
	return &SFTPFile{file: f, name: name, client: fs.client}, nil
}

// OpenFile opens a file on the remote host with the specified flags and permissions.
func (fs *SFTPFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	f, err := fs.client.OpenFile(name, flag)
	if err != nil {
		return nil, err
	}
	// Set permissions after creation if needed
	if flag&os.O_CREATE != 0 {
		_ = fs.client.Chmod(name, perm)
	}
	return &SFTPFile{file: f, name: name, client: fs.client}, nil
}

// Remove removes a file on the remote host.
func (fs *SFTPFs) Remove(name string) error {
	return fs.client.Remove(name)
}

// RemoveAll removes a path and all children on the remote host.
func (fs *SFTPFs) RemoveAll(path string) error {
	// Check if it's a directory
	info, err := fs.client.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fs.client.Remove(path)
	}

	// Walk the directory tree in reverse to remove children first
	walker := fs.client.Walk(path)
	var paths []string
	for walker.Step() {
		if walker.Err() != nil {
			continue
		}
		paths = append(paths, walker.Path())
	}
	// Remove in reverse order (deepest first)
	var errs []string
	for i := len(paths) - 1; i >= 0; i-- {
		info, err := fs.client.Stat(paths[i])
		if err != nil {
			errs = append(errs, fmt.Sprintf("stat %s: %v", paths[i], err))
			continue
		}
		if info.IsDir() {
			if err := fs.client.RemoveDirectory(paths[i]); err != nil {
				errs = append(errs, fmt.Sprintf("rmdir %s: %v", paths[i], err))
			}
		} else {
			if err := fs.client.Remove(paths[i]); err != nil {
				errs = append(errs, fmt.Sprintf("rm %s: %v", paths[i], err))
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors during RemoveAll: %s", strings.Join(errs, "; "))
	}
	return nil
}

// Rename renames a file on the remote host.
func (fs *SFTPFs) Rename(oldname, newname string) error {
	return fs.client.Rename(oldname, newname)
}

// Stat returns file info for a path on the remote host.
func (fs *SFTPFs) Stat(name string) (os.FileInfo, error) {
	return fs.client.Stat(name)
}

// Name returns the name of this filesystem.
func (fs *SFTPFs) Name() string {
	return "SFTPFs"
}

// Chmod changes permissions of a file on the remote host.
func (fs *SFTPFs) Chmod(name string, mode os.FileMode) error {
	return fs.client.Chmod(name, mode)
}

// Chown changes ownership of a file on the remote host.
func (fs *SFTPFs) Chown(name string, uid, gid int) error {
	return fs.client.Chown(name, uid, gid)
}

// Chtimes changes the access and modification times of a file.
func (fs *SFTPFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return fs.client.Chtimes(name, atime, mtime)
}

// SFTPFile wraps an sftp.File to implement afero.File.
type SFTPFile struct {
	file   *sftp.File
	name   string
	client *sftp.Client
}

// Close closes the file.
func (f *SFTPFile) Close() error {
	return f.file.Close()
}

// Read reads from the file.
func (f *SFTPFile) Read(p []byte) (int, error) {
	return f.file.Read(p)
}

// ReadAt reads from the file at the given offset.
func (f *SFTPFile) ReadAt(p []byte, off int64) (int, error) {
	return f.file.ReadAt(p, off)
}

// Seek seeks to the given position.
func (f *SFTPFile) Seek(offset int64, whence int) (int64, error) {
	return f.file.Seek(offset, whence)
}

// Write writes to the file.
func (f *SFTPFile) Write(p []byte) (int, error) {
	return f.file.Write(p)
}

// WriteAt writes to the file at the given offset.
func (f *SFTPFile) WriteAt(p []byte, off int64) (int, error) {
	return f.file.WriteAt(p, off)
}

// Name returns the name of the file.
func (f *SFTPFile) Name() string {
	return f.name
}

// Readdir reads the directory contents.
func (f *SFTPFile) Readdir(count int) ([]os.FileInfo, error) {
	entries, err := f.client.ReadDir(f.name)
	if err != nil {
		return nil, err
	}
	if count > 0 && count < len(entries) {
		entries = entries[:count]
	}
	return entries, nil
}

// Readdirnames reads and returns directory entry names.
func (f *SFTPFile) Readdirnames(n int) ([]string, error) {
	entries, err := f.Readdir(n)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	return names, nil
}

// Stat returns file info.
func (f *SFTPFile) Stat() (os.FileInfo, error) {
	return f.file.Stat()
}

// Sync is a no-op for SFTP (writes are synchronous).
func (f *SFTPFile) Sync() error {
	return nil
}

// Truncate truncates the file to the given size.
func (f *SFTPFile) Truncate(size int64) error {
	return f.file.Truncate(size)
}

// WriteString writes a string to the file.
func (f *SFTPFile) WriteString(s string) (int, error) {
	return f.file.Write([]byte(s))
}
