package ssh

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	"github.com/rs/zerolog/log"
)

// fileTransfer handles file transfer operations via SFTP.
type fileTransfer struct {
	client *SSHClient
	config *Config
}

// UploadFile uploads a single file to the remote host via SFTP.
func (c *SSHClient) UploadFile(ctx context.Context, localPath string, remotePath string, mode uint32) error {
	if c.fileTransfer == nil {
		return &TransportError{
			Op:          "upload",
			Err:         fmt.Errorf("file transfer not initialized"),
			IsTemporary: false,
			IsAuthError: false,
		}
	}
	return c.fileTransfer.uploadFile(ctx, localPath, remotePath, mode)
}

// DownloadFile downloads a single file from the remote host via SFTP.
func (c *SSHClient) DownloadFile(ctx context.Context, remotePath string, localPath string) error {
	if c.fileTransfer == nil {
		return &TransportError{
			Op:          "download",
			Err:         fmt.Errorf("file transfer not initialized"),
			IsTemporary: false,
			IsAuthError: false,
		}
	}
	return c.fileTransfer.downloadFile(ctx, remotePath, localPath)
}

// UploadDirectory recursively uploads a directory to the remote host.
func (c *SSHClient) UploadDirectory(ctx context.Context, localPath string, remotePath string) error {
	if c.fileTransfer == nil {
		return &TransportError{
			Op:          "upload-dir",
			Err:         fmt.Errorf("file transfer not initialized"),
			IsTemporary: false,
			IsAuthError: false,
		}
	}
	return c.fileTransfer.uploadDirectory(ctx, localPath, remotePath)
}

// DownloadDirectory recursively downloads a directory from the remote host.
func (c *SSHClient) DownloadDirectory(ctx context.Context, remotePath string, localPath string) error {
	if c.fileTransfer == nil {
		return &TransportError{
			Op:          "download-dir",
			Err:         fmt.Errorf("file transfer not initialized"),
			IsTemporary: false,
			IsAuthError: false,
		}
	}
	return c.fileTransfer.downloadDirectory(ctx, remotePath, localPath)
}

// SetFilePermissions sets file permissions on the remote host.
func (c *SSHClient) SetFilePermissions(ctx context.Context, remotePath string, mode uint32) error {
	if c.fileTransfer == nil {
		return &TransportError{
			Op:          "chmod",
			Err:         fmt.Errorf("file transfer not initialized"),
			IsTemporary: false,
			IsAuthError: false,
		}
	}
	return c.fileTransfer.setPermissions(ctx, remotePath, mode)
}

// SetFileOwnership sets file ownership on the remote host.
func (c *SSHClient) SetFileOwnership(ctx context.Context, remotePath string, uid int, gid int) error {
	if c.fileTransfer == nil {
		return &TransportError{
			Op:          "chown",
			Err:         fmt.Errorf("file transfer not initialized"),
			IsTemporary: false,
			IsAuthError: false,
		}
	}
	return c.fileTransfer.setOwnership(ctx, remotePath, uid, gid)
}

// ComputeChecksum calculates the checksum of a remote file.
func (c *SSHClient) ComputeChecksum(ctx context.Context, remotePath string) (string, error) {
	if c.fileTransfer == nil {
		return "", &TransportError{
			Op:          "checksum",
			Err:         fmt.Errorf("file transfer not initialized"),
			IsTemporary: false,
			IsAuthError: false,
		}
	}
	return c.fileTransfer.computeChecksum(ctx, remotePath)
}

// createSFTPClient creates a new SFTP client.
func (f *fileTransfer) createSFTPClient() (*sftp.Client, error) {
	sshClient, err := f.client.getClient()
	if err != nil {
		return nil, err
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return nil, &TransportError{
			Op:          "sftp-init",
			Err:         fmt.Errorf("failed to create SFTP client: %w", err),
			IsTemporary: true,
			IsAuthError: false,
		}
	}

	return sftpClient, nil
}

// uploadFile uploads a single file to the remote host.
func (f *fileTransfer) uploadFile(ctx context.Context, localPath string, remotePath string, mode uint32) error {
	startTime := time.Now()

	log.Debug().
		Str("local", localPath).
		Str("remote", remotePath).
		Uint32("mode", mode).
		Msg("uploading file")

	// Open the local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return &TransportError{
			Op:          "upload",
			Err:         fmt.Errorf("failed to open local file: %w", err),
			IsTemporary: false,
			IsAuthError: false,
		}
	}
	defer localFile.Close()

	// Get file info for size
	fileInfo, err := localFile.Stat()
	if err != nil {
		return &TransportError{
			Op:          "upload",
			Err:         fmt.Errorf("failed to stat local file: %w", err),
			IsTemporary: false,
			IsAuthError: false,
		}
	}

	// Create SFTP client
	sftpClient, err := f.createSFTPClient()
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	// Ensure remote directory exists
	remoteDir := filepath.Dir(remotePath)
	if err := sftpClient.MkdirAll(remoteDir); err != nil {
		return &TransportError{
			Op:          "upload",
			Err:         fmt.Errorf("failed to create remote directory: %w", err),
			IsTemporary: false,
			IsAuthError: false,
		}
	}

	// Create the remote file
	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return &TransportError{
			Op:          "upload",
			Err:         fmt.Errorf("failed to create remote file: %w", err),
			IsTemporary: true,
			IsAuthError: false,
		}
	}
	defer remoteFile.Close()

	// Copy the file with context awareness
	bytesWritten, err := f.copyWithContext(ctx, remoteFile, localFile)
	if err != nil {
		return &TransportError{
			Op:          "upload",
			Err:         fmt.Errorf("failed to copy file: %w", err),
			IsTemporary: true,
			IsAuthError: false,
		}
	}

	// Set file permissions if specified
	if mode > 0 {
		if err := sftpClient.Chmod(remotePath, os.FileMode(mode)); err != nil {
			log.Warn().Err(err).Msg("failed to set file permissions")
		}
	}

	duration := time.Since(startTime)

	log.Info().
		Str("local", localPath).
		Str("remote", remotePath).
		Int64("bytes", bytesWritten).
		Int64("size", fileInfo.Size()).
		Dur("duration", duration).
		Msg("file uploaded successfully")

	return nil
}

// downloadFile downloads a single file from the remote host.
func (f *fileTransfer) downloadFile(ctx context.Context, remotePath string, localPath string) error {
	startTime := time.Now()

	log.Debug().
		Str("remote", remotePath).
		Str("local", localPath).
		Msg("downloading file")

	// Create SFTP client
	sftpClient, err := f.createSFTPClient()
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	// Open the remote file
	remoteFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return &TransportError{
			Op:          "download",
			Err:         fmt.Errorf("failed to open remote file: %w", err),
			IsTemporary: true,
			IsAuthError: false,
		}
	}
	defer remoteFile.Close()

	// Get file info for size
	fileInfo, err := remoteFile.Stat()
	if err != nil {
		return &TransportError{
			Op:          "download",
			Err:         fmt.Errorf("failed to stat remote file: %w", err),
			IsTemporary: true,
			IsAuthError: false,
		}
	}

	// Ensure local directory exists
	localDir := filepath.Dir(localPath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return &TransportError{
			Op:          "download",
			Err:         fmt.Errorf("failed to create local directory: %w", err),
			IsTemporary: false,
			IsAuthError: false,
		}
	}

	// Create the local file
	localFile, err := os.Create(localPath)
	if err != nil {
		return &TransportError{
			Op:          "download",
			Err:         fmt.Errorf("failed to create local file: %w", err),
			IsTemporary: false,
			IsAuthError: false,
		}
	}
	defer localFile.Close()

	// Copy the file with context awareness
	bytesWritten, err := f.copyWithContext(ctx, localFile, remoteFile)
	if err != nil {
		return &TransportError{
			Op:          "download",
			Err:         fmt.Errorf("failed to copy file: %w", err),
			IsTemporary: true,
			IsAuthError: false,
		}
	}

	duration := time.Since(startTime)

	log.Info().
		Str("remote", remotePath).
		Str("local", localPath).
		Int64("bytes", bytesWritten).
		Int64("size", fileInfo.Size()).
		Dur("duration", duration).
		Msg("file downloaded successfully")

	return nil
}

// uploadDirectory recursively uploads a directory.
func (f *fileTransfer) uploadDirectory(ctx context.Context, localPath string, remotePath string) error {
	log.Debug().
		Str("local", localPath).
		Str("remote", remotePath).
		Msg("uploading directory")

	// Create SFTP client
	sftpClient, err := f.createSFTPClient()
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	// Walk the local directory
	return filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(remotePath, relPath)

		if info.IsDir() {
			// Create remote directory
			log.Debug().Str("dir", targetPath).Msg("creating remote directory")
			if err := sftpClient.MkdirAll(targetPath); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
		} else {
			// Upload file
			if err := f.uploadFile(ctx, path, targetPath, uint32(info.Mode().Perm())); err != nil {
				return fmt.Errorf("failed to upload file %s: %w", path, err)
			}
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		return nil
	})
}

// downloadDirectory recursively downloads a directory.
func (f *fileTransfer) downloadDirectory(ctx context.Context, remotePath string, localPath string) error {
	log.Debug().
		Str("remote", remotePath).
		Str("local", localPath).
		Msg("downloading directory")

	// Create SFTP client
	sftpClient, err := f.createSFTPClient()
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	// Walk the remote directory
	walker := sftpClient.Walk(remotePath)
	for walker.Step() {
		if err := walker.Err(); err != nil {
			return &TransportError{
				Op:          "download-dir",
				Err:         fmt.Errorf("failed to walk remote directory: %w", err),
				IsTemporary: true,
				IsAuthError: false,
			}
		}

		// Calculate relative path
		relPath, err := filepath.Rel(remotePath, walker.Path())
		if err != nil {
			return err
		}

		targetPath := filepath.Join(localPath, relPath)

		if walker.Stat().IsDir() {
			// Create local directory
			log.Debug().Str("dir", targetPath).Msg("creating local directory")
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
		} else {
			// Download file
			if err := f.downloadFile(ctx, walker.Path(), targetPath); err != nil {
				return fmt.Errorf("failed to download file %s: %w", walker.Path(), err)
			}
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	return nil
}

// setPermissions sets file permissions on the remote host.
func (f *fileTransfer) setPermissions(ctx context.Context, remotePath string, mode uint32) error {
	log.Debug().
		Str("path", remotePath).
		Uint32("mode", mode).
		Msg("setting file permissions")

	// Create SFTP client
	sftpClient, err := f.createSFTPClient()
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	if err := sftpClient.Chmod(remotePath, os.FileMode(mode)); err != nil {
		return &TransportError{
			Op:          "chmod",
			Err:         fmt.Errorf("failed to set permissions: %w", err),
			IsTemporary: false,
			IsAuthError: false,
		}
	}

	log.Info().Str("path", remotePath).Uint32("mode", mode).Msg("permissions set successfully")
	return nil
}

// setOwnership sets file ownership on the remote host.
func (f *fileTransfer) setOwnership(ctx context.Context, remotePath string, uid int, gid int) error {
	log.Debug().
		Str("path", remotePath).
		Int("uid", uid).
		Int("gid", gid).
		Msg("setting file ownership")

	// Use executor to run chown command (requires sudo)
	cmd := fmt.Sprintf("chown %d:%d %s", uid, gid, remotePath)
	_, stderr, err := f.client.ExecuteCommandWithSudo(ctx, cmd, "")
	if err != nil {
		return &TransportError{
			Op:          "chown",
			Err:         fmt.Errorf("failed to set ownership: %s", stderr),
			IsTemporary: false,
			IsAuthError: false,
		}
	}

	log.Info().
		Str("path", remotePath).
		Int("uid", uid).
		Int("gid", gid).
		Msg("ownership set successfully")

	return nil
}

// computeChecksum calculates the SHA256 checksum of a remote file.
func (f *fileTransfer) computeChecksum(ctx context.Context, remotePath string) (string, error) {
	log.Debug().Str("path", remotePath).Msg("computing checksum")

	// Use executor to run sha256sum command
	cmd := fmt.Sprintf("sha256sum %s", remotePath)
	stdout, stderr, err := f.client.ExecuteCommand(ctx, cmd)
	if err != nil {
		return "", &TransportError{
			Op:          "checksum",
			Err:         fmt.Errorf("failed to compute checksum: %s", stderr),
			IsTemporary: false,
			IsAuthError: false,
		}
	}

	// Parse the output (format: "checksum  filename")
	parts := splitChecksum(stdout)
	if len(parts) < 1 {
		return "", &TransportError{
			Op:          "checksum",
			Err:         fmt.Errorf("invalid checksum output: %s", stdout),
			IsTemporary: false,
			IsAuthError: false,
		}
	}

	checksum := parts[0]
	log.Debug().Str("path", remotePath).Str("checksum", checksum).Msg("checksum computed")

	return checksum, nil
}

// computeLocalChecksum calculates the SHA256 checksum of a local file.
func (f *fileTransfer) computeLocalChecksum(localPath string) (string, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// copyWithContext copies data from src to dst while respecting context cancellation.
func (f *fileTransfer) copyWithContext(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	// Use a buffer for efficient copying
	buf := make([]byte, 32*1024) // 32KB buffer
	var written int64

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return written, ctx.Err()
		default:
		}

		// Read from source
		nr, err := src.Read(buf)
		if nr > 0 {
			// Write to destination
			nw, err := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if err != nil {
				return written, err
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return written, err
		}
	}

	return written, nil
}

// splitChecksum splits checksum output into parts.
func splitChecksum(output string) []string {
	// Split by whitespace
	parts := make([]string, 0)
	current := ""
	for _, ch := range output {
		if ch == ' ' || ch == '\t' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
