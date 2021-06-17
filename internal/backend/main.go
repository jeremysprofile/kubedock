package backend

import (
	"io"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

// Backend is the interface to orchestrate and manage kubernetes objects.
type Backend interface {
	StartContainer(*types.Container) (DeployState, error)
	CreatePortForwards(*types.Container)
	CreateReverseProxies(*types.Container)
	GetServiceClusterIP(*types.Container) (string, error)
	DeleteAll() error
	DeleteWithKubedockID(string) error
	DeleteContainer(*types.Container) error
	DeleteContainersOlderThan(time.Duration) error
	DeleteServicesOlderThan(time.Duration) error
	CopyToContainer(*types.Container, []byte, string) error
	ExecContainer(*types.Container, *types.Exec, io.Writer) (int, error)
	GetLogs(*types.Container, bool, int, chan struct{}, io.Writer) error
	GetImageExposedPorts(string) (map[string]struct{}, error)
}

// instance is the internal representation of the Backend object.
type instance struct {
	cli       kubernetes.Interface
	cfg       *rest.Config
	initImage string
	namespace string
	timeOut   int
}

// Config is the structure to instantiate a Backend object
type Config struct {
	// Client is the kubernetes clientset
	Client kubernetes.Interface
	// RestConfig is the kubernetes config
	RestConfig *rest.Config
	// Namespace is the namespace in which all actions are performed
	Namespace string
	// InitImage is the image that is used as init container to prepare vols
	InitImage string
	// TimeOut is the max amount of time to wait until a container started
	TimeOut time.Duration
}

// New will return an Backend instance.
func New(cfg Config) Backend {
	return &instance{
		cli:       cfg.Client,
		cfg:       cfg.RestConfig,
		initImage: cfg.InitImage,
		namespace: cfg.Namespace,
		timeOut:   int(cfg.TimeOut.Seconds()),
	}
}
