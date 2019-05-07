package gosftp

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SFTP struct {
	client         *sftp.Client
	host           string
	port           int
	protocol       string
	user, password string
}

func New(user, password, host, protocol, pathToKey string, port int) (*SFTP, error) {
	sftp := &SFTP{user: user, password: password, protocol: protocol, port: port, host: host}
	if pathToKey != "" {
		if key, err := readKey(pathToKey); err == nil {
			sftp.connectSSH(key)
		}
	}
	return sftp, nil
}
func (s *SFTP) connectSSH(keyBuffer []byte) error {
	key, err := ssh.ParsePrivateKey(keyBuffer)
	if err != nil {
		return err
	}
	auths := []ssh.AuthMethod{ssh.PublicKeys(key)}
	config := &ssh.ClientConfig{User: s.user, Auth: auths, HostKeyCallback: ssh.InsecureIgnoreHostKey(), HostKeyAlgorithms: []string{"ssh-dss"}}
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%v", s.host, s.port), config)
	if err != nil {
		return err
	}
	client, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	s.client = client
	return nil
}
func readKey(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}
func (s *SFTP) GetFile(folderPath, fileName string) (*sftp.File, error) {
	return s.client.Open(path.Join(folderPath, fileName))
}
func (s *SFTP) DeleteFile(folderPath, fileName string) error {
	return s.client.Remove(path.Join(folderPath, fileName))
}
func (s *SFTP) GetFiles(filesPath string) (map[string]*sftp.File, error) {
	list, err := s.client.ReadDir(filesPath)
	if err != nil {
		return nil, err
	}
	files := make(map[string]*sftp.File)
	for _, item := range list {
		if file, err := s.GetFile(filesPath, item.Name()); err == nil {
			files[item.Name()] = file
		} else {
			return nil, err
		}
	}
	return files, nil
}
func SaveFilesToLocalFolder(files map[string]*sftp.File, folderPath string) error {
	for name, file := range files {
		dstFile, err := os.Create(path.Join(folderPath, name))
		if err != nil {
			return err
		}
		_, err = io.Copy(dstFile, file)
		if err != nil {
			return err
		}
		err = dstFile.Sync()
		if err != nil {
			return err
		}
		dstFile.Close()
	}
	return nil
}
