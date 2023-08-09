package chroma_dapr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	chroma "github.com/amikos-tech/chroma-go"
	openai "github.com/amikos-tech/chroma-go/openai"
	"github.com/dapr/components-contrib/bindings"
	"github.com/dapr/kit/logger"
	"strconv"
	"time"
)

const (
	reset                                 bindings.OperationKind = "reset"
	version                               bindings.OperationKind = "version"
	heartbeat                             bindings.OperationKind = "heartbeat"
	createCollection                      bindings.OperationKind = "createCollection"
	deleteCollection                      bindings.OperationKind = "deleteCollection"
	listCollections                       bindings.OperationKind = "listCollections"
	getCollection                         bindings.OperationKind = "getCollection"
	openAIEmbeddingFunction               string                 = "openai"
	cohereEmbeddingFunction               string                 = "cohere"
	sentenceTransformersEmbeddingFunction string                 = "sentenceTransformers"
	urlProperty                           string                 = "url"
	openAIApiKeyProperty                  string                 = "openAIApiKey"
	cohereApiKeyProperty                  string                 = "cohereApiKey"
	huggingFaceApiKeyProperty             string                 = "huggingFaceApiKey"
	defaultEmbeddingFunctionProperty      string                 = "defaultEmbeddingFunction"
)

type chromaMetadata struct {
	Url                      string
	OperationTimeout         time.Duration
	OpeAIAPIKey              string
	CohereAPIKey             string
	HuggingFaceAPIKey        string
	DefaultEmbeddingFunction string // openai, cohere, sentenceTransformers
}

type ChromaBindingComponent struct {
	client *chroma.Client
	meta   chromaMetadata
	logger logger.Logger
}

func NewChroma(logger logger.Logger) bindings.OutputBinding {
	return &ChromaBindingComponent{
		logger: logger,
	}
}

func (c *ChromaBindingComponent) Init(ctx context.Context, meta bindings.Metadata) (err error) {
	// Called to initialize the component with its configured metadata...
	c.meta = chromaMetadata{
		Url:                      meta.Properties[urlProperty],
		OpeAIAPIKey:              meta.Properties[openAIApiKeyProperty],
		CohereAPIKey:             meta.Properties[cohereApiKeyProperty],
		HuggingFaceAPIKey:        meta.Properties[huggingFaceApiKeyProperty],
		DefaultEmbeddingFunction: meta.Properties[defaultEmbeddingFunctionProperty],
	}
	if c.meta.Url == "" {
		return errors.New("missing host field from metadata")
	}
	//check if 	DefaultEmbeddingFunction string // openai, cohere, sentenceTransformers
	if c.meta.DefaultEmbeddingFunction != "" {
		if c.meta.DefaultEmbeddingFunction != openAIEmbeddingFunction &&
			c.meta.DefaultEmbeddingFunction != cohereEmbeddingFunction &&
			c.meta.DefaultEmbeddingFunction != sentenceTransformersEmbeddingFunction {
			return errors.New("incorrect defaultEmbeddingFunction field from metadata")
		}
	} else {
		c.meta.DefaultEmbeddingFunction = openAIEmbeddingFunction
	}

	if val, ok := meta.Properties["operationTimeout"]; ok && val != "" {
		c.meta.OperationTimeout, err = time.ParseDuration(val)
		if err != nil {
			return errors.New("incorrect operationTimeout field from metadata")
		}
	}
	c.logger = logger.NewLogger("dapr.binding.chroma")
	c.client = chroma.NewClient(c.meta.Url)
	return nil
}

func (c *ChromaBindingComponent) GetComponentMetadata() map[string]string {
	return map[string]string{
		"url":                      c.meta.Url,
		"operationTimeout":         c.meta.OperationTimeout.String(),
		"defaultEmbeddingFunction": c.meta.DefaultEmbeddingFunction,
	}
}

type createCollectionRequest struct {
	Name              string                 `json:"name"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	GetOrCreate       bool                   `json:"getOrCreate,omitempty"`
	EmbeddingFunction string                 `json:"embeddingFunction,omitempty"`
	DistanceFunction  string                 `json:"distanceFunction,omitempty"`
}

func validateCreateCollectionRequest(req createCollectionRequest) error {
	if req.Name == "" {
		return errors.New("missing name field from request")
	}
	if req.EmbeddingFunction != openAIEmbeddingFunction &&
		req.EmbeddingFunction != cohereEmbeddingFunction &&
		req.EmbeddingFunction != sentenceTransformersEmbeddingFunction {
		return errors.New("incorrect embeddingFunction field from request")
	}
	if req.DistanceFunction != string(chroma.L2) &&
		req.DistanceFunction != string(chroma.IP) &&
		req.DistanceFunction != string(chroma.COSINE) {
		return errors.New("incorrect distanceFunction field from request")
	}
	return nil
}

func (c *ChromaBindingComponent) getEmbeddingFunction(embeddingFunctionString string) (chroma.EmbeddingFunction, error) {
	switch embeddingFunctionString {
	case openAIEmbeddingFunction:
		if c.meta.OpeAIAPIKey == "" {
			return nil, errors.New("missing openAIAPIKey field from metadata")
		}
		return openai.NewOpenAIEmbeddingFunction(c.meta.OpeAIAPIKey), nil
	case cohereEmbeddingFunction:
		return nil, errors.New("cohere embedding function not implemented yet")
	case sentenceTransformersEmbeddingFunction:
		return nil, errors.New("sentenceTransformers embedding function not implemented yet")
	default:
		return nil, errors.New("invalid embedding function")
	}
}

func (c *ChromaBindingComponent) getDistanceFunction(distanceFunctionString string) (chroma.DistanceFunction, error) {

	switch distanceFunctionString {
	case string(chroma.L2):
		return chroma.L2, nil
	case string(chroma.IP):
		return chroma.IP, nil
	case string(chroma.COSINE):
		return chroma.COSINE, nil
	default:
		return chroma.L2, errors.New("invalid distance function")
	}
}

func (c *ChromaBindingComponent) Invoke(ctx context.Context, req *bindings.InvokeRequest) (resp *bindings.InvokeResponse, err error) {
	startTime := time.Now().UTC()
	resp = &bindings.InvokeResponse{
		Metadata: map[string]string{
			"operation":  string(req.Operation),
			"start-time": startTime.Format(time.RFC3339Nano),
		},
	}
	switch req.Operation { //nolint:exhaustive
	case reset:
		r, err := c.client.Reset()
		if err != nil {
			return nil, err
		}
		resp.Data = []byte(strconv.FormatBool(r))

	case version:
		c.logger.Debugf("getting version")
		v, err := c.client.Version()
		if err != nil {
			return nil, err
		}
		resp.Data = []byte(v)

	case heartbeat:
		c.logger.Debugf("getting version")
		h, err := c.client.Heartbeat()
		if err != nil {
			return nil, err
		}
		d, err := json.Marshal(h)
		if err != nil {
			return nil, err
		}
		resp.Data = d

	case createCollection:
		var createCollectionReq createCollectionRequest
		err := json.Unmarshal(req.Data, &createCollectionReq)
		if err != nil {
			return nil, err
		}
		err = validateCreateCollectionRequest(createCollectionReq)
		if err != nil {
			return nil, err
		}
		var embeddingFunction chroma.EmbeddingFunction
		c.logger.Infof("createCollection: %v", createCollectionReq)
		embeddingFunction, err = c.getEmbeddingFunction(createCollectionReq.EmbeddingFunction)
		if err != nil {
			return nil, err
		}

		c.logger.Infof("distanceFunctionString: %s", createCollectionReq.DistanceFunction)
		distanceFunction, err := c.getDistanceFunction(createCollectionReq.DistanceFunction)
		if err != nil {
			return nil, err
		}
		newCollection, err := c.client.CreateCollection(createCollectionReq.Name, createCollectionReq.Metadata, createCollectionReq.GetOrCreate, embeddingFunction, distanceFunction)
		if err != nil {
			return nil, err
		}
		d, err := json.Marshal(newCollection)
		if err != nil {
			return nil, err
		}
		resp.Data = d

	case deleteCollection:
		var deleteCollectionReq struct {
			Name string `json:"name"`
		}
		err := json.Unmarshal(req.Data, &deleteCollectionReq)
		if err != nil {
			return nil, err
		}
		deletedCollection, err := c.client.DeleteCollection(deleteCollectionReq.Name)
		if err != nil {
			return nil, err
		}
		d, err := json.Marshal(deletedCollection)
		if err != nil {
			return nil, err
		}
		resp.Data = d

	case listCollections:
		collections, err := c.client.ListCollections()
		if err != nil {
			return nil, err
		}
		d, err := json.Marshal(collections)
		if err != nil {
			return nil, err
		}
		resp.Data = d

	case getCollection:
		var getCollectionReq struct {
			Name string `json:"name"`
		}
		err := json.Unmarshal(req.Data, &getCollectionReq)
		if err != nil {
			return nil, err
		}
		collection, err := c.client.GetCollection(getCollectionReq.Name, nil)
		if err != nil {
			return nil, err
		}
		d, err := json.Marshal(collection)
		if err != nil {
			return nil, err
		}
		resp.Data = d
	default:
		return nil, fmt.Errorf(
			"invalid operation type: %s. Expected %v",
			req.Operation, []bindings.OperationKind{reset, version, heartbeat, createCollection, deleteCollection, listCollections, getCollection},
		)
	}
	endTime := time.Now().UTC()
	resp.Metadata["end-time"] = endTime.Format(time.RFC3339Nano)
	resp.Metadata["duration"] = endTime.Sub(startTime).String()

	return resp, nil
}

func (c *ChromaBindingComponent) Operations() []bindings.OperationKind {
	return []bindings.OperationKind{
		reset,
		version,
		heartbeat,
		createCollection,
		deleteCollection,
		listCollections,
		getCollection,
	}
}
