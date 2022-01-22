package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	pb "github.com/micro/micro/v3/proto/auth"
	auth "github.com/micro/micro/v3/service/auth"
	client "github.com/micro/micro/v3/service/client"
	log "github.com/micro/micro/v3/service/logger"
	store "github.com/micro/micro/v3/service/store"
	authns "github.com/micro/micro/v3/util/auth/namespace"
	namespace "github.com/micro/micro/v3/util/namespace"
	errors "github.com/pkg/errors"

	orgservice "org-service/proto"
)

type OrgService struct{}

type OrgValue struct {
	Created string `json:"created_at" yaml:"created_at"`
}

const NAMESPACE_PREFIX = "orgs"
const SEPARATOR = "."
const SEPARATOR_REPLACEMENT = "-"

func Sanitize(name string) string {
	return strings.Replace(name, SEPARATOR, SEPARATOR_REPLACEMENT, -1)
}

func Qid(req *orgservice.Request) string {
	return fmt.Sprintf("%s.%s", NAMESPACE_PREFIX, Sanitize(req.Name))
}

// Create a org
func (e *OrgService) Create(ctx context.Context, req *orgservice.Request, rsp *orgservice.Response) error {
	log.Info("Received OrgService.Create request")

	account, ok := auth.AccountFromContext(ctx)
	if !ok {
		return errors.New("could not get account")
	}

	log.Info("ACCOUNT ID: %s", account.ID)
	log.Info("ACCOUNT SECRET: %s", account.Name)

	if req.Options == nil {
		req.Options = &orgservice.Options{}
	}

	if len(req.Options.Namespace) == 0 {
		req.Options.Namespace = namespace.FromContext(ctx)
	}

	err := authns.AuthorizeAdmin(ctx, req.Options.Namespace, "org.Org.Create")
	if err != nil {
		return errors.Wrapf(err, "could not authorize request")
	}

	ac := pb.NewAccountsService("auth", client.DefaultClient)
	resp, err := ac.List(ctx, &pb.ListAccountsRequest{})
	if err != nil {
		return errors.Wrapf(err, "could not call accounts list")
	}

	accountjson, err := json.Marshal(resp)
	if err != nil {
		return errors.Wrapf(err, "error marshalling generated account to json")
	}

	qid := Qid(req)

	records, err := store.Read(qid)
	if err != nil && err != store.ErrNotFound {
		return errors.Wrapf(err, "error reading from store for key '%s'", qid)
	}

	if len(records) != 0 {
		return errors.New(fmt.Sprintf("org already exists: %s", qid))
	}

	val := &OrgValue{
		Created: fmt.Sprintf("%d", time.Now().Unix()),
	}

	valbytes, err := json.Marshal(val)
	if err != nil {
		return errors.Wrapf(err, "error marshalling org value to json")
	}

	rec := &store.Record{
		Key:   qid,
		Value: valbytes,
	}

	err = store.Write(rec)
	if err != nil {
		return errors.Wrapf(err, "error writing to store with key '%s'", req.Name)
	}

	rsp.Msg = fmt.Sprintf("org "+req.Name+" created. - %s", accountjson)
	rsp.Created = val.Created
	return nil
}

// Delete a org
func (e *OrgService) Delete(ctx context.Context, req *orgservice.Request, rsp *orgservice.Response) error {
	log.Info("Received OrgService.Delete request")

	qid := Qid(req)

	err := store.Delete(qid)
	if err != nil {
		return errors.Wrapf(err, "error deleting org '%s'", qid)
	}

	rsp.Msg = "org " + req.Name + " deleted."
	return nil
}

// List a org
func (e *OrgService) List(ctx context.Context, req *orgservice.Request, rsp *orgservice.Response) error {
	log.Info("Received OrgService.List request")

	records, err := store.List(store.ListPrefix(NAMESPACE_PREFIX))
	if err != nil {
		return errors.Wrapf(err, "error listing orgs")
	}

	data, err := json.Marshal(records)
	if err != nil {
		return errors.Wrapf(err, "error marshalling orgs to json")
	}

	rsp.Msg = string(data)
	return nil
}

func NewOrgService() *OrgService {
	return &OrgService{}
}
