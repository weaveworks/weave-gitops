package validate

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/sourceignore"
	"github.com/yannh/kubeconform/pkg/output"
	"github.com/yannh/kubeconform/pkg/resource"
	"github.com/yannh/kubeconform/pkg/validator"
)

func Validate(log logger.Logger, targetDir, rootDir, kubernetesVersion, fluxVersion string) error {
	var (
		o     output.Output
		err   error
		files []string
	)

	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}

	fluxSchemaDir := filepath.Join(userCacheDir, ".gitops", "flux", "schemas")
	if _, err := os.Stat(fluxSchemaDir); os.IsNotExist(err) {
		if err := os.MkdirAll(fluxSchemaDir, 0o755); err != nil {
			return err
		}

		cli := cleanhttp.DefaultClient()
		url := fmt.Sprintf("https://github.com/fluxcd/flux2/releases/download/%s/crd-schemas.tar.gz", fluxVersion)
		response, err := cli.Get(url)

		if err != nil {
			return err
		}

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Println(err)
			}
		}(response.Body)

		const schemaStrictPrefix = "master-standalone-strict"

		if err := untar(filepath.Join(fluxSchemaDir, schemaStrictPrefix), response.Body); err != nil {
			return err
		}

		ksConfig, err := cli.Get("https://json.schemastore.org/kustomization.json")
		if err != nil {
			return err
		}

		defer func(body io.ReadCloser) {
			if err := body.Close(); err != nil {
				fmt.Println(err)
			}
		}(ksConfig.Body)

		ksConfigFile, err := os.Create(filepath.Join(fluxSchemaDir, schemaStrictPrefix, "kustomize.config.k8s.io-kustomization-kustomize-v1beta1.json"))
		if err != nil {
			return err
		}

		defer func(file *os.File) {
			if err := file.Close(); err != nil {
				fmt.Println(err)
			}
		}(ksConfigFile)

		if _, err := io.Copy(ksConfigFile, ksConfig.Body); err != nil {
			return err
		}
	}

	// load sourceignore patterns
	ignorePath := filepath.Join(rootDir, sourceignore.IgnoreFilename)

	ps, err := sourceignore.ReadIgnoreFile(ignorePath, nil)
	if err != nil {
		log.Warningf("Couldn't read the ignore file %s: %v", ignorePath, err)
	}

	filter := sourceignore.IgnoreFileFilter(ps, []string{})

	// walk the target directory and find all YAML files
	err = filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml") && !filter(path, info) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return err
	}

	if o, err = output.New("text", false, false, false); err != nil {
		return err
	}

	cacheDir := filepath.Join(userCacheDir, ".gitops", "schema-cache")
	// make sure the cache directory exists
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return err
	}

	schemaLocations := []string{
		// special case for K8s Kustomization config
		fluxSchemaDir + "/master-standalone{{ .StrictSuffix }}/{{ .Group }}-{{ .ResourceKind }}{{ .KindSuffix }}.json",
		// standard Flux schemas
		fluxSchemaDir,
		// the default K8s schemas
		"default",
	}

	var v validator.Validator
	v, err = validator.New(schemaLocations, validator.Opts{
		Cache:                cacheDir,
		Debug:                false,
		SkipTLS:              false,
		SkipKinds:            nil,
		RejectKinds:          nil,
		KubernetesVersion:    kubernetesVersion,
		Strict:               true,
		IgnoreMissingSchemas: true,
	})

	if err != nil {
		return err
	}

	validationResults := make(chan validator.Result)
	ctx, cancel := context.WithCancel(context.Background())
	successChan := processResults(cancel, o, validationResults, true)

	var (
		resourcesChan          <-chan resource.Resource
		errors                 <-chan error
		ignoreFilenamePatterns []string
	)

	resourcesChan, errors = resource.FromFiles(ctx, files, ignoreFilenamePatterns)

	// Process discovered resources across multiple workers
	const numberOfWorkers = 4

	wg := sync.WaitGroup{}
	for i := 0; i < numberOfWorkers; i++ {
		wg.Add(1)

		go func(resources <-chan resource.Resource, validationResults chan<- validator.Result, v validator.Validator) {
			for res := range resources {
				validationResults <- v.ValidateResource(res)
			}

			wg.Done()
		}(resourcesChan, validationResults, v)
	}

	wg.Add(1)

	go func() {
		// Process errors while discovering resources
		for err := range errors {
			if err == nil {
				continue
			}

			if err, ok := err.(resource.DiscoveryError); ok {
				validationResults <- validator.Result{
					Resource: resource.Resource{Path: err.Path},
					Err:      err.Err,
					Status:   validator.Error,
				}
			} else {
				validationResults <- validator.Result{
					Resource: resource.Resource{},
					Err:      err,
					Status:   validator.Error,
				}
			}

			cancel()
		}

		wg.Done()
	}()

	wg.Wait()

	close(validationResults)

	success := <-successChan

	if err := o.Flush(); err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("validation failed")
	}

	return nil
}

func processResults(cancel context.CancelFunc, o output.Output, validationResults <-chan validator.Result, exitOnError bool) <-chan bool {
	success := true
	result := make(chan bool)

	go func() {
		for res := range validationResults {
			if res.Status == validator.Error || res.Status == validator.Invalid {
				success = false
			}

			if o != nil {
				if err := o.Write(res); err != nil {
					fmt.Fprint(os.Stderr, "failed writing log\n")
				}
			}

			if !success && exitOnError {
				cancel() // early exit - signal to stop searching for resources
				break
			}
		}

		for range validationResults { // allow resource finders to exit
		}

		result <- success
	}()

	return result
}

func untar(destDir string, r io.Reader) (retErr error) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}

	defer func(gzr *gzip.Reader) {
		err := gzr.Close()
		if err != nil {
			retErr = err
		}
	}(gzr)

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {
		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(destDir, header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {
		// if it's a dir and doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0o755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			if err := f.Close(); err != nil {
				return err
			}
		}
	}
}
