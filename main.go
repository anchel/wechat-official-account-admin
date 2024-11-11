package main

import (
	"context"
	"embed"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/anchel/wechat-official-account-admin/controllers" // Import all controllers
	"github.com/anchel/wechat-official-account-admin/lib/types"
	"github.com/anchel/wechat-official-account-admin/modules/weixin"
	"github.com/anchel/wechat-official-account-admin/mongodb"
	"github.com/anchel/wechat-official-account-admin/routes"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	redislib "github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "woaa",
	Short: "woaa",
	Long:  "wechat-official-account-admin, server and frontend",
	Run: func(cmd *cobra.Command, args []string) {
		err := run()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

//go:embed wechat-official-account-admin-fe/dist
var frontend embed.FS

func run() error {
	log.Println("starting run")

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
		return err
	}

	log.Println("starting InitMongoDB")
	mongoClient, err := mongodb.InitMongoDB()
	if err != nil {
		log.Println("Error mongodb.InitMongoDB")
		return err
	}
	// 定期检查mongo数据库健康状态
	go func() {
		for {
			time.Sleep(30 * time.Second)
			if !mongoClient.HealthCheck(context.Background()) {
				log.Println("Trying to reconnect to MongoDB...")
				if err := mongoClient.Reconnect(context.Background()); err != nil {
					log.Fatalf("Failed to reconnect to MongoDB: %v", err)
				}
			}
		}
	}()

	// 优雅退出
	defer mongoClient.Disconnect(context.Background())

	r := gin.Default()
	err = r.SetTrustedProxies(nil)
	if err != nil {
		log.Println("Error r.SetTrustedProxies")
		return err
	}

	store, err := redis.NewStore(3, "tcp", os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PASSWORD"), []byte("secret666"))
	if err != nil {
		log.Println("Error redis.NewStore")
		return err
	}
	r.Use(sessions.Sessions("mysession", store))

	r.Use(static.Serve("/", static.EmbedFolder(frontend, "wechat-official-account-admin-fe/dist")))

	exePath, err := os.Executable()
	if err != nil {
		log.Println("Error os.Executable")
		return err
	}
	wd := filepath.Dir(exePath)
	r.Static("/files/upload-image", filepath.Join(wd, "files", "upload-image"))           // 上传的图片
	r.Static("/files/wx-download-media", filepath.Join(wd, "files", "wx-download-media")) // 下载的微信素材

	// enable single page application
	r.NoRoute(func(c *gin.Context) {
		fmt.Println("NoRoute", c.Request.URL.Path)
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusOK, gin.H{
				"code":    1,
				"message": "api not found",
			})
			return
		}
		c.FileFromFS("/wechat-official-account-admin-fe/dist/template.html", http.FS(frontend))
	})

	excludeLoginPaths := []string{
		"/api/system/user/register",
		"/api/system/user/login",
		"/api/system/user/userinfo",
	}

	group := r.Group("/api", func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if lo.Contains(excludeLoginPaths, path) {
			// log.Println("check login path exclude", path)
			ctx.Next()
		} else {
			session := sessions.Default(ctx)
			login, ok := session.Get("login").(int)
			if !ok || login != 1 {
				// log.Println("check login false")
				ctx.JSON(http.StatusOK, gin.H{
					"code":    1,
					"message": "no login",
				})
				ctx.Abort()
			} else {
				// log.Println("check login success")
				ctx.Next()
			}
		}
	}, func(ctx *gin.Context) {
		path := ctx.Request.URL.Path

		// 先排除不需要登录态的
		if lo.Contains(excludeLoginPaths, path) {
			ctx.Next()
			return
		}

		// 再排除不需要appid的
		if strings.HasPrefix(path, "/api/system/") {
			ctx.Next()
			return
		}

		session := sessions.Default(ctx)
		appidInfo := session.Get("appid")
		if appidInfo == nil {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    1,
				"message": "no appid",
			})
			ctx.Abort()
		} else {
			app, ok := appidInfo.(types.SessionAppidInfo)
			if ok {
				ctx.Set("appid", app.AppID)
				ctx.Next()
			} else {
				ctx.JSON(http.StatusOK, gin.H{
					"code":    1,
					"message": "no appid",
				})
				ctx.Abort()
			}
		}
	})
	routes.InitRoutes(group)

	rdb := redislib.NewClient(&redislib.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
	})

	gob.Register(types.SessionAppidInfo{})

	err = weixin.InitWeixin(rdb)
	if err != nil {
		log.Println("Error weixin.InitWeixin")
		return err
	}

	wxmp := r.Group("/wxmp")
	wxmp.GET("/:appid/handler", weixin.Serve)
	wxmp.POST("/:appid/handler", weixin.Serve)

	return r.Run(os.Getenv("LISTEN_ADDR"))
}
