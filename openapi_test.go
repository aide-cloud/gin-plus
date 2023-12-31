package ginplus

import (
	"context"
	"log"
	"testing"

	"github.com/gin-gonic/gin"
)

type (
	Api struct {
	}

	ApiDetailReq struct {
		Id uint `uri:"id" desc:"数据自增唯一ID"`
	}

	ApiDetailResp struct {
		Id     uint   `json:"id"`
		Name   string `json:"name"`
		Remark string `json:"remark"`
	}

	ApiListReq struct {
		Current int    `form:"current"`
		Size    int    `form:"size"`
		Keyword string `form:"keyword"`
	}
	ApiListResp struct {
		Total int64          `json:"total"`
		List  []*ApiInfoItem `json:"list"`
	}

	Extent struct {
		A `json:"a"`
		B `json:"b"`
	}

	ApiInfoItem struct {
		Name   string `json:"name"`
		Id     uint   `json:"id"`
		Remark string `json:"remark"`
		Extent Extent `json:"extent"`
	}

	ApiUpdateReq struct {
		Id     uint   `uri:"id"`
		Name   string `json:"name"`
		Remark string `json:"remark"`
	}
	ApiUpdateResp struct {
		Id uint `json:"id"`
	}

	DelApiReq struct {
		Id uint `uri:"id"`
	}

	DelApiResp struct {
		Id uint `json:"id"`
	}
)

func (l *Api) GetDetail(_ context.Context, req *ApiDetailReq) (*ApiDetailResp, error) {
	log.Println("Api.GetDetail")
	return &ApiDetailResp{
		Id:     req.Id,
		Name:   "demo",
		Remark: "hello world",
	}, nil
}

func (l *Api) GetList(_ context.Context, _ *ApiListReq) (*ApiListResp, error) {
	log.Println("Api.GetList")
	return &ApiListResp{
		Total: 100,
		List: []*ApiInfoItem{
			{
				Id:     10,
				Name:   "demo",
				Remark: "hello world",
			},
		},
	}, nil
}

func (l *Api) UpdateInfo(_ context.Context, req *ApiUpdateReq) (*ApiUpdateResp, error) {
	log.Println("Api.UpdateInfo")
	return &ApiUpdateResp{Id: req.Id}, nil
}

func (l *Api) DeleteInfo(_ context.Context, req *DelApiReq) (*DelApiResp, error) {
	log.Println("Api.DeleteInfo")
	return &DelApiResp{Id: req.Id}, nil
}

func TestGinEngine_execute(t *testing.T) {
	New(gin.Default(),
		WithControllers(&Api{}),
		WithOpenApiYaml("./", "openapi.yaml"),
		AppendHttpMethodPrefixes(
			HttpMethod{
				Prefix: "Update",
				Method: Put,
			}, HttpMethod{
				Prefix: "Del",
				Method: Delete,
			},
		),
	)
}
