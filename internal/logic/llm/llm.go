package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/grpool"

	"llama-proxy/internal/middleware"
	"llama-proxy/internal/model"
)

type sLLM struct {
	mu               sync.RWMutex
	currentModel     string
	currentModelAddr string
	cmd              *exec.Cmd
	worker           *grpool.Pool
	lastAccessTime   int64
	idleTimer        *time.Timer
}

func newLLM() *sLLM {
	return &sLLM{
		worker: grpool.New(1), // 单worker确保串行处理
	}
}

func (s *sLLM) ModelList(ctx context.Context) []*model.Model {
	modelConfigs, err := loadModelsFromConfig(ctx)
	if err != nil {
		g.Log().Error(ctx, "Failed to load models from config:", err)
		return nil
	}

	models := make([]*model.Model, 0, len(modelConfigs))
	for _, cfg := range modelConfigs {
		models = append(models, &model.Model{
			Id:     cfg.Id,
			Object: "model",
		})
	}
	return models
}

func (s *sLLM) ModelDetail(ctx context.Context, modelId string) (*model.Model, error) {
	modelConfigs, err := loadModelsFromConfig(ctx)
	if err != nil {
		return nil, err
	}

	for _, cfg := range modelConfigs {
		if cfg.Id == modelId {
			return &model.Model{
				Id:     cfg.Id,
				Object: "model",
			}, nil
		}
	}
	return nil, nil
}

func loadModelsFromConfig(ctx context.Context) ([]*model.ModelConfig, error) {
	var configs []*model.ModelConfig
	err := g.Cfg().MustGet(ctx, "openai.models").Scan(&configs)
	if err != nil {
		return nil, err
	}
	return configs, nil
}

func getModelConfigById(ctx context.Context, modelId string) (*model.ModelConfig, error) {
	configs, err := loadModelsFromConfig(ctx)
	if err != nil {
		return nil, err
	}

	for _, cfg := range configs {
		if cfg.Id == modelId {
			return cfg, nil
		}
	}
	return nil, nil
}

func getDefaultParams(ctx context.Context) *model.DefaultParams {
	params := &model.DefaultParams{}
	err := g.Cfg().MustGet(ctx, "openai.default_params").Scan(params)
	if err != nil {
		g.Log().Warning(ctx, "Failed to load default params, using defaults")
	}
	return params
}

func getFreePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func buildCommandArgs(cfg *model.ModelConfig, defaultParams *model.DefaultParams, port int) []string {
	args := []string{}

	args = append(args, "--model", cfg.ModelPath)
	args = append(args, "--port", strconv.Itoa(port))

	if cfg.CtxSize != 0 {
		args = append(args, "--ctx-size", strconv.Itoa(cfg.CtxSize))
	} else if defaultParams.CtxSize != 0 {
		args = append(args, "--ctx-size", strconv.Itoa(defaultParams.CtxSize))
	}

	if cfg.Threads != 0 {
		args = append(args, "--threads", strconv.Itoa(cfg.Threads))
	} else if defaultParams.Threads != 0 {
		args = append(args, "--threads", strconv.Itoa(defaultParams.Threads))
	}

	if cfg.Parallel != 0 {
		args = append(args, "--parallel", strconv.Itoa(cfg.Parallel))
	} else if defaultParams.Parallel != 0 {
		args = append(args, "--parallel", strconv.Itoa(defaultParams.Parallel))
	}

	if cfg.GpuLayers != 0 {
		args = append(args, "--n-gpu-layers", strconv.Itoa(cfg.GpuLayers))
	} else if defaultParams.GpuLayers != 0 {
		args = append(args, "--n-gpu-layers", strconv.Itoa(defaultParams.GpuLayers))
	}

	if cfg.CacheTypeK != "" {
		args = append(args, "--cache-type-k", cfg.CacheTypeK)
	} else if defaultParams.CacheTypeK != "" {
		args = append(args, "--cache-type-k", defaultParams.CacheTypeK)
	}

	if cfg.CacheTypeV != "" {
		args = append(args, "--cache-type-v", cfg.CacheTypeV)
	} else if defaultParams.CacheTypeV != "" {
		args = append(args, "--cache-type-v", defaultParams.CacheTypeV)
	}

	if cfg.FlashAttn != "" {
		if cfg.FlashAttn == "on" {
			args = append(args, "--flash-attn")
		}
	} else if defaultParams.FlashAttn != "" && defaultParams.FlashAttn == "on" {
		args = append(args, "--flash-attn")
	}

	if cfg.NoMmap || defaultParams.NoMmap {
		args = append(args, "--no-mmap")
	}

	if cfg.Mlock || defaultParams.Mlock {
		args = append(args, "--mlock")
	}

	if cfg.Temp != 0 {
		args = append(args, "--temp", strconv.FormatFloat(cfg.Temp, 'f', -1, 64))
	} else if defaultParams.Temp != 0 {
		args = append(args, "--temp", strconv.FormatFloat(defaultParams.Temp, 'f', -1, 64))
	}

	if cfg.TopP != 0 {
		args = append(args, "--top-p", strconv.FormatFloat(cfg.TopP, 'f', -1, 64))
	} else if defaultParams.TopP != 0 {
		args = append(args, "--top-p", strconv.FormatFloat(defaultParams.TopP, 'f', -1, 64))
	}

	if cfg.TopK != 0 {
		args = append(args, "--top-k", strconv.Itoa(cfg.TopK))
	} else if defaultParams.TopK != 0 {
		args = append(args, "--top-k", strconv.Itoa(defaultParams.TopK))
	}

	if cfg.MinP != 0 {
		args = append(args, "--min-p", strconv.FormatFloat(cfg.MinP, 'f', -1, 64))
	} else if defaultParams.MinP != 0 {
		args = append(args, "--min-p", strconv.FormatFloat(defaultParams.MinP, 'f', -1, 64))
	}

	if cfg.PresencePenalty != 0 {
		args = append(args, "--presence-penalty", strconv.FormatFloat(cfg.PresencePenalty, 'f', -1, 64))
	} else if defaultParams.PresencePenalty != 0 {
		args = append(args, "--presence-penalty", strconv.FormatFloat(defaultParams.PresencePenalty, 'f', -1, 64))
	}

	if cfg.RepeatPenalty != 0 {
		args = append(args, "--repeat-penalty", strconv.FormatFloat(cfg.RepeatPenalty, 'f', -1, 64))
	} else if defaultParams.RepeatPenalty != 0 {
		args = append(args, "--repeat-penalty", strconv.FormatFloat(defaultParams.RepeatPenalty, 'f', -1, 64))
	}

	if cfg.VisionModelPath != "" {
		args = append(args, "--mmproj", cfg.VisionModelPath)
	}

	return args
}

// unloadCurrentModelLocked 卸载当前模型（调用方必须持有写锁）
func (s *sLLM) unloadCurrentModelLocked(ctx context.Context) {
	if s.idleTimer != nil {
		s.idleTimer.Stop()
		s.idleTimer = nil
	}
	if s.cmd != nil {
		_ = s.cmd.Process.Kill()
		_ = s.cmd.Wait()
	}
	s.currentModel = ""
	s.currentModelAddr = ""
	s.cmd = nil
}

// loadModelLocked 加载模型（调用方必须持有写锁）
func (s *sLLM) loadModelLocked(ctx context.Context, modelId string) (addr string, err error) {
	cfg, err := getModelConfigById(ctx, modelId)
	if err != nil {
		return "", err
	}

	defaultParams := getDefaultParams(ctx)

	port, err := getFreePort()
	if err != nil {
		return "", err
	}

	args := buildCommandArgs(cfg, defaultParams, port)

	llamaPath := g.Cfg().MustGet(ctx, "openai.llama_server_path", "llama-server").String()

	s.cmd = exec.Command(llamaPath, args...)
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr

	if err = s.cmd.Start(); err != nil {
		return "", err
	}

	s.currentModel = modelId
	s.currentModelAddr = fmt.Sprintf("http://127.0.0.1:%d", port)

	go s.monitorProcess(ctx, modelId)

	s.resetIdleTimer(ctx)

	g.Log().Infof(ctx, "Model loaded: %s, address: %s", modelId, s.currentModelAddr)
	return s.currentModelAddr, nil
}

func (s *sLLM) monitorProcess(ctx context.Context, modelId string) {
	if s.cmd == nil {
		return
	}

	err := s.cmd.Wait()
	if err != nil {
		g.Log().Warningf(ctx, "llama-server process for model %s exited with error: %v", modelId, err)
	} else {
		g.Log().Infof(ctx, "llama-server process for model %s exited normally", modelId)
	}

	s.mu.Lock()
	if s.currentModel == modelId {
		g.Log().Infof(ctx, "Model %s process exited, resetting model state", modelId)
		s.currentModel = ""
		s.currentModelAddr = ""
		s.cmd = nil
	}
	s.mu.Unlock()
}

func (s *sLLM) resetIdleTimer(ctx context.Context) {
	if s.idleTimer != nil {
		s.idleTimer.Stop()
	}

	idleTimeout := g.Cfg().MustGet(ctx, "openai.idle_timeout", 0).Int()
	if idleTimeout <= 0 {
		return
	}

	s.idleTimer = time.AfterFunc(time.Duration(idleTimeout)*time.Second, func() {
		s.mu.Lock()
		modelId := s.currentModel
		s.mu.Unlock()

		if modelId != "" {
			g.Log().Infof(ctx, "Model %s idle timeout reached, unloading...", modelId)
			s.UnloadModel(ctx, modelId)
		}
	})
}

func (s *sLLM) touch(ctx context.Context) {
	s.mu.Lock()
	s.lastAccessTime = time.Now().Unix()
	s.mu.Unlock()
	s.resetIdleTimer(ctx)
}

func (s *sLLM) LoadModel(ctx context.Context, modelId string) (addr string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentModel == modelId && s.currentModelAddr != "" {
		return s.currentModelAddr, nil
	}

	if s.currentModel != "" {
		g.Log().Infof(ctx, "Switching model from %s to %s", s.currentModel, modelId)
		s.unloadCurrentModelLocked(ctx)
	}

	return s.loadModelLocked(ctx, modelId)
}

func (s *sLLM) UnloadModel(ctx context.Context, modelId string) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentModel != modelId {
		return gerror.NewCode(gcode.CodeInvalidParameter, "Model is not loaded")
	}

	g.Log().Infof(ctx, "Model unloaded: %s", modelId)
	s.unloadCurrentModelLocked(ctx)
	return nil
}

func (s *sLLM) IsModelLoaded(ctx context.Context, modelId string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentModel == modelId, nil
}

// 需要模型参数的接口路径
var modelRequiredPaths = map[string]bool{
	"/v1/chat/completions": true,
	"/v1/responses":        true,
	"/v1/embeddings":       true,
}

func (s *sLLM) ServeHTTP(ctx context.Context, r *ghttp.Request) {
	path := r.Request.URL.Path
	c := make(chan struct{})
	defer func() { <-c }()

	s.worker.Add(ctx, func(ctx context.Context) {
		defer func() {
			c <- struct{}{}
		}()

		// 检查是否需要模型参数
		needsModel := modelRequiredPaths[path]
		modelId := ""

		if needsModel {
			// 需要模型参数的接口
			body := r.GetBodyString()
			if body != "" {
				var params map[string]interface{}
				if err := json.Unmarshal([]byte(body), &params); err == nil {
					if m, ok := params["model"].(string); ok {
						modelId = m
					}
				}
			}

			if modelId == "" {
				middleware.WriteError(r, http.StatusBadRequest, "model parameter is required", "invalid_request_error", "missing_model")
				return
			}

			// 需要检查和加载模型
			s.mu.RLock()
			for {
				isLoaded := s.currentModel == modelId
				if !isLoaded {
					s.mu.RUnlock()
					s.mu.Lock()

					if s.currentModel != "" && s.currentModel != modelId {
						g.Log().Infof(ctx, "Switching model from %s to %s", s.currentModel, modelId)
						s.unloadCurrentModelLocked(ctx)
					}

					if s.currentModel == modelId {
						s.mu.Unlock()
						s.mu.RLock()
						continue
					}

					_, err := s.loadModelLocked(ctx, modelId)
					if err != nil {
						s.mu.Unlock()
						g.Log().Errorf(ctx, "Failed to load model %s: %v", modelId, err)
						middleware.WriteError(r, http.StatusInternalServerError, "Failed to load model: "+err.Error(), "server_error", "model_load_failed")
						return
					}

					s.mu.Unlock()
					s.mu.RLock()
					g.Log().Infof(ctx, "Model %s loaded successfully", modelId)
				} else {
					break
				}
			}
		}

		// 获取后端地址
		s.mu.RLock()
		backendAddr := s.currentModelAddr
		s.mu.RUnlock()

		if backendAddr == "" {
			middleware.WriteError(r, http.StatusServiceUnavailable, "No model is loaded, please make a request to a model-required endpoint first", "service_unavailable", "no_model_loaded")
			return
		}

		s.touch(ctx)

		// 执行反向代理
		backend, err := url.Parse(backendAddr)
		if err != nil {
			g.Log().Error(ctx, "Failed to parse backend URL:", err)
			middleware.WriteError(r, http.StatusInternalServerError, "Failed to parse backend URL", "server_error", "invalid_backend_url")
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(backend)
		proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
			g.Log().Error(ctx, "Proxy error:", err)
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(`{"error": {"message": "Bad gateway", "type": "server_error", "param": null, "code": "proxy_error"}}`))
		}

		r.Header.Set("Host", backend.Host)
		proxy.ServeHTTP(r.Response.Writer, r.Request)
	})
}
