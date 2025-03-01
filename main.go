package main

import (
	"context"
	"embed"
	"encoding/gob"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/anchel/wechat-official-account-admin/controllers" // Import all controllers
	"github.com/anchel/wechat-official-account-admin/lib/logger"
	"github.com/anchel/wechat-official-account-admin/lib/types"
	"github.com/anchel/wechat-official-account-admin/lib/utils"
	"github.com/anchel/wechat-official-account-admin/modules/weixin"
	"github.com/anchel/wechat-official-account-admin/mongodb"
	"github.com/anchel/wechat-official-account-admin/routes"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-contrib/static"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"

	"github.com/joho/godotenv"
	redislib "github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var rootCmd = &cobra.Command{
	Use:   "woaa",
	Short: "woaa",
	Long:  "wechat-official-account-admin, server and frontend",
	Run: func(cmd *cobra.Command, args []string) {
		err := run()
		if err != nil {
			logger.Error(err.Error())
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
	// 初始化日志
	logger.SetDefault(logger.NewLogger(logger.NewTerminalHandlerWithLevel(os.Stderr, logger.LevelDebug, true)))

	logger.Info("starting run")

	if utils.CheckEnvFile() {
		logger.Info("loading .env file")
		err := godotenv.Load()
		if err != nil {
			logger.Error("Error loading .env file")
			return err
		}
	}

	logger.Info("starting InitMongoDB")
	mongoClient, err := mongodb.InitMongoDB()
	if err != nil {
		logger.Error("Error mongodb.InitMongoDB")
		return err
	}
	// 定期检查mongo数据库健康状态
	go func() {
		for {
			time.Sleep(30 * time.Second)
			if !mongoClient.HealthCheck(context.Background()) {
				logger.Warn("Trying to reconnect to MongoDB...")
				if err := mongoClient.Reconnect(context.Background()); err != nil {
					logger.Crit("Failed to reconnect to MongoDB", err.Error())
				}
			}
		}
	}()

	// 优雅退出
	defer mongoClient.Disconnect(context.Background())

	// create gin instance
	r := gin.New()
	err = r.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		logger.Error("Error r.SetTrustedProxies")
		return err
	}
	r.Use(gin.Logger())

	excludeLogPaths := []string{
		"/api/system/user/userinfo",
		"/api/appid/session_info",
		"/api/request-log/list",
	}
	zcore := logger.NewMongoZapCore[types.GinRequestLogInfo](zap.DebugLevel, func() (*mongo.Collection, error) {
		return mongoClient.GetCollection("request-logs")
	})
	defer zcore.Sync()

	loggerMongo := zap.New(zcore)

	r.Use(ginzap.GinzapWithConfig(loggerMongo, &ginzap.Config{
		TimeFormat:   time.RFC3339,
		UTC:          true,
		DefaultLevel: zap.InfoLevel,
		Skipper: func(c *gin.Context) bool {
			result := true

			if strings.HasPrefix(c.Request.URL.Path, "/api/") && !lo.Contains(excludeLogPaths, c.Request.URL.Path) {
				result = false
			}

			return result
		},
		Context: ginzap.Fn(func(c *gin.Context) []zapcore.Field {
			fields := []zapcore.Field{}
			session := sessions.Default(c)
			appidInfo := session.Get("appid")
			if appidInfo != nil {
				app, ok := appidInfo.(types.SessionAppidInfo)
				if ok {
					fields = append(fields, zap.String("appid", app.AppID))
				}
			}
			return fields
		}),
	}))

	// store request logs in mongodb
	r.Use(ginzap.RecoveryWithZap(loggerMongo, true))

	// enable session
	store, err := redis.NewStore(3, "tcp", os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PASSWORD"), []byte("secret666"))
	if err != nil {
		logger.Error("Error redis.NewStore")
		return err
	}
	r.Use(sessions.Sessions("mysession", store))

	// serve frontend assets, such as html,css,image and so on
	r.Use(static.Serve("/", static.EmbedFolder(frontend, "wechat-official-account-admin-fe/dist")))

	exePath, err := os.Executable()
	if err != nil {
		logger.Error("Error os.Executable")
		return err
	}
	wd := filepath.Dir(exePath)
	r.Static("/files/upload-image", filepath.Join(wd, "files", "upload-image"))           // 上传的图片
	r.Static("/files/wx-download-media", filepath.Join(wd, "files", "wx-download-media")) // 下载的微信素材

	// enable single page application
	r.NoRoute(func(c *gin.Context) {
		logger.Info("NoRoute", "path", c.Request.URL.Path)
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusOK, gin.H{
				"code":    1,
				"message": "api not found",
			})
			return
		}
		accept := c.Request.Header.Get("Accept")
		if !strings.Contains(accept, "html") {
			c.JSON(http.StatusOK, gin.H{
				"code":    1,
				"message": "not html",
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

	// check login status
	// check whether the apis that must work with the official account carries the appid
	group := r.Group("/api", func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if lo.Contains(excludeLoginPaths, path) {
			// logger.Info("check login path exclude", path)
			ctx.Next()
		} else {
			session := sessions.Default(ctx)
			login, ok := session.Get("login").(int)
			if !ok || login != 1 {
				// logger.Info("check login false")
				ctx.JSON(http.StatusOK, gin.H{
					"code":    1,
					"message": "no login",
				})
				ctx.Abort()
			} else {
				// logger.Info("check login success")
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

	// for common redis operation
	rdb := redislib.NewClient(&redislib.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
	})

	gob.Register(types.SessionAppidInfo{})

	// init weixin module
	err = weixin.InitWeixin(rdb)
	if err != nil {
		logger.Error("Error weixin.InitWeixin")
		return err
	}

	wxmp := r.Group("/wxmp")
	wxmp.GET("/:appid/handler", weixin.Serve)
	wxmp.POST("/:appid/handler", weixin.Serve)

	// server start listening
	return r.Run(os.Getenv("LISTEN_ADDR"))
}
