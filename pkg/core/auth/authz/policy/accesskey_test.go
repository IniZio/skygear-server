package policy

import (
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDenyNoAccessKey(t *testing.T) {
	Convey("Test DenyNoAccessKey", t, func() {
		Convey("should return error if auth context has no access key", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{
				mAccessKeyType: model.NoAccessKey,
			}

			err := DenyNoAccessKey(req, ctx)
			So(err, ShouldNotBeEmpty)
		})

		Convey("should not return error if auth context has api key ", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{
				mAccessKeyType: model.APIAccessKey,
			}

			err := DenyNoAccessKey(req, ctx)
			So(err, ShouldBeEmpty)
		})

		Convey("should not return error if auth context has master key ", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{
				mAccessKeyType: model.MasterAccessKey,
			}

			err := DenyNoAccessKey(req, ctx)
			So(err, ShouldBeEmpty)
		})
	})
}

func TestRequireMasterKey(t *testing.T) {
	Convey("Test RequireMasterKey", t, func() {
		Convey("should return error if auth context has no access key", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{
				mAccessKeyType: model.NoAccessKey,
			}

			err := RequireMasterKey(req, ctx)
			So(err, ShouldNotBeEmpty)
		})

		Convey("should return error if auth context has api key ", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{
				mAccessKeyType: model.APIAccessKey,
			}

			err := RequireMasterKey(req, ctx)
			So(err, ShouldNotBeEmpty)
		})

		Convey("should not return error if auth context has master key ", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{
				mAccessKeyType: model.MasterAccessKey,
			}

			err := RequireMasterKey(req, ctx)
			So(err, ShouldBeEmpty)
		})
	})
}
