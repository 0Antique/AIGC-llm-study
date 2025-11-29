

# AIGC-llm-study

黑马程序员大模型课程[[B站链接]](https://www.bilibili.com/video/BV1sMk6BtE3f?spm_id_from=333.788.videopod.episodes&vd_source=affb8bc5c4837c4659502cba93c353cd)

学习完成时间：2025.11.29  21:34

学习内容包含：

1. AIGC的简介——文生图领域
2. 常见的图像生成模型——VAE,GAN,扩散模型
3. Stable Diffusion模型的架构
4. 云平台HAI和模型的训练——Dreabooth方法和LoRA方法
5. 混元接口和小程序实现文生图

主要学习收获：

1. 了解了几种图像生成算法
   * VAE变分自编码器
   * GAN对抗生成网络
   * 扩散模型
2. 学习了Stable Diffusion的训练方法
   * Dreambooth学习风格
   * LoRA进行微调训练
3. 搭建C端的文生图**应用**
   * 使用小程序前端
   * 后端使用Java调用腾讯混元API返回BASE64
   *  使用腾讯云对象存储 (COS）