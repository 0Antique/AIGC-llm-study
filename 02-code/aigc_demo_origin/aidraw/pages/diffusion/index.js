// index.js
// 获取应用实例
const app = getApp()

Page({
  data: {
    totalTask: 30,
    queueLen: 0,
    leftTime: 40,
    beginTime: 0,
    processTime: 0,
    taskStatus: "STOP",
    inputValue: "",
    tags: [],
    option: [],
    loading: false,
    index: 0,
    motto: 'Hello World',
    userInfo: {},
    hasUserInfo: false,
    canIUse: wx.canIUse('button.open-type.getUserInfo'),
    canIUseGetUserProfile: false,
    canIUseOpenData: wx.canIUse('open-data.type.userAvatarUrl') && wx.canIUse('open-data.type.userNickName') // 如需尝试获取用户信息可改为false
  },
  // 事件处理函数
  bindViewTap() {
    wx.navigateTo({
      url: '../logs/logs'
    })
  },
  onLoad() {
    if (wx.getUserProfile) {
      this.setData({
        canIUseGetUserProfile: true
      })
    }
    this.onTimeout();
  },
 
  getUserProfile(e) { 
    // 推荐使用wx.getUserProfile获取用户信息，开发者每次通过该接口获取用户个人信息均需用户确认，开发者妥善保管用户快速填写的头像昵称，避免重复弹窗
    wx.getUserProfile({
      desc: '展示用户信息', // 声明获取用户个人信息后的用途，后续会展示在弹窗中，请谨慎填写
      success: (res) => {
        console.log(res)
        this.setData({
          userInfo: res.userInfo,
          hasUserInfo: true
        })
      }
    })
  },
  getUserInfo(e) {
    // 不推荐使用getUserInfo获取用户信息，预计自2021年4月13日起，getUserInfo将不再弹出弹窗，并直接返回匿名的用户个人信息
    console.log(e)
    this.setData({
      userInfo: e.detail.userInfo,
      hasUserInfo: true
    })
  },

  facefusion(modelUrl, imageUrl) {
    var that = this;
    that.setData({
      taskStatus: "融合中...",
      processTime: (Date.parse(new Date()) - that.data.beginTime)/ 1000
    })
    wx.request({
      url: 'http://127.0.0.1:8000/frontend/fusion',
      data: {
        "session_id": "123",
        "model_url": modelUrl,
        "image_url": imageUrl
      },
      method: "POST",
      header: {
        'Content-Type': "application/json"
      },
      success (res) {
        if (res.data == null) {
          wx.showToast({
            icon: "error",
            title: '请求融合失败',
          })
          return
        }
        
        if (res.data.result_url !== "") {
          console.log("draw image: ", res.data.result_url)
          that.drawInputImage('#input_canvas_fusion', res.data.result_url);
          that.setData({
            Resp: {},
            taskStatus: "STOP",
            queueLen: 0,
            loading: false,
          })
          // clearTimeout(that.data.ticker);
        } else {
          that.setData({
            taskStatus: "PROCESSING",
            processTime: (Date.parse(new Date()) - that.data.beginTime)/ 1000
          })
        }
        // a portrait of an old coal miner in 19th century, beautiful painting with highly detailed face by greg rutkowski and magali villanueve
      },
      fail(res) {
        wx.showToast({
          icon: "error",
          title: '请求融合失败',
        })
        console.log(res)
      }
    })
  },

  enentloop() {
    var that = this
    if (!that.data.Resp || !that.data.Resp.job_id) {
      // console.log("not found jobid")
      return
    }
    return new Promise(function(yes, no) {
      wx.request({
      url: 'http://127.0.0.1:8000/frontend/query',
      data: {
        "session_id": "123",
        "job_id": that.data.Resp.job_id
      },
      method: "POST",
      header: {
        'Content-Type': "application/json"
      },
      success (res) {
        yes("hello");
        if (res.data == null) {
          wx.showToast({
            icon: "error",
            title: '请求查询失败',
          })
          return
        }
        console.log(Date.parse(new Date()), res.data)
        that.setData({
          Job: res.data,
        })
        console.log("job_status: ", res.data.job_status)
        if (res.data.job_status === "FINISNED") {
          console.log("AI image: ", res.data.result_url)
          that.setData({
            loading: false
          })
          that.drawInputImage('#input_canvas', res.data.result_url);
          console.log(888, res.data.result_url)
          if (!that.data.inputUrl || that.data.inputUrl.length === 0) {
            that.setData({
              Resp: {},
              taskStatus: "STOP",
              queueLen: 0,
              loading: false,
            })
          } else {
            that.setData({
              Resp: {},
            })
            // 请求融合
            that.facefusion(res.data.result_url, that.data.inputUrl);
          }
        } else {
          that.setData({
            taskStatus: "PROCESSING",
            processTime: (Date.parse(new Date()) - that.data.beginTime)/ 1000
          })
        }
      },
      fail(res) {
        wx.showToast({
          icon: "error",
          title: '请求查询失败',
        })
        console.log(res)
      }
    })
  })
  },

  onTimeout:  function() {
    // 开启定时器
    var that = this;
    let ticker = setTimeout(async function() {
      await that.enentloop();
      that.onTimeout();
    }, 1 * 1000); // 毫秒数
    // clearTimeout(ticker);
    that.setData({
      ticker: ticker
    });
  },

  imageDraw() {
    var that = this
    var opt = {}
    if (that.data.option && that.data.option.length > 0) {
      opt = {
        "tags": that.data.option
      }
    }
    console.log("option:", opt)
    wx.request({
      url: 'http://127.0.0.1:8000/frontend/create',
      data: {
        "prompt": that.data.inputValue
      },
      method: "POST",
      header: {
        'Content-Type': "application/json"
      },
      success (res) {
        if (res.data == null) {
          wx.showToast({
            icon: "error",
            title: '请求失败',
          })
          return
        }
        console.log(res.data)
        // let raw = JSON.parse(res.data)
        that.setData({
          queueLen: res.data.queue_len,
          leftTime: res.data.remain_time,
          Resp: res.data,
        })
        that.setData({
          totalTask: res.data.total_cnt,
          beginTime: Date.parse(new Date())
        })
      },
      fail(res) {
        wx.showToast({
          icon: "error",
          title: '请求失败',
        })
        that.setData({
          loading: false
        })
      }
    })
  },

  drawInputImage: function(canvasItem, url) {
    var that = this;
    console.log("result_url: ", url)

    let resUrl = url; // that.data.Job.result_url;
    
    wx.downloadFile({
      url: resUrl,
      success: function(res) {
        var imagePath = res.tempFilePath
        wx.getImageInfo({
          src: imagePath,
          success: function(res) {
            wx.createSelectorQuery()
            .select(canvasItem) // 在 WXML 中填入的 id
            .fields({ node: true, size: true })
            .exec((r) => {
              // Canvas 对象
              const canvas = r[0].node
              // 渲染上下文
              const ctx = canvas.getContext('2d')
              // Canvas 画布的实际绘制宽高 
              const width = r[0].width
              const height = r[0].height
              // 初始化画布大小
              const dpr = wx.getWindowInfo().pixelRatio
              canvas.width = width * dpr
              canvas.height = height * dpr
              ctx.scale(dpr, dpr)
              ctx.clearRect(0, 0, width, height)

              let radio = height / res.height
              console.log("radio:", radio)
              const img = canvas.createImage()
              var x = width / 2 - (res.width * radio / 2)

              img.src = imagePath
              img.onload = function() {
                ctx.drawImage(img, x, 0, res.width * radio, res.height * radio)
              }
            })
          }
        })
      }
    })
  },

  handlerInput(e) {
    this.setData({
      inputValue: e.detail.value
    })
  },

  handlerSearch(e) {
    console.log("input: ", this.data.inputValue)

    if (this.data.inputValue.length == 0) {
      wx.showToast({
        icon: "error",
        title: '请输入你的创意 ',
      })
      return
    }
    this.setData({
      loading: true,
    })
    this.imageDraw()
  },
  handlerInputPos(e) {
    console.log(e)
    this.setData({
      inputValue: e.detail.value
    })
  },
  handlerInputFusion(e) {
    console.log(e)
    this.setData({
      inputUrl: e.detail.value
    })
  },
  handlerInputImage(e) {
    console.log(e)
  },
  clickItem(e) {
    let $bean = e.currentTarget.dataset
    console.log(e)
    console.log("value: ", $bean.bean)
    this.setData({
      option: $bean.bean,
    })
    console.log('clickItem: ', this.data.loading)
    this.imageDraw()
  }
})