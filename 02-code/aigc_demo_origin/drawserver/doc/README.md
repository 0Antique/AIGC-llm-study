### 示例1：  
a portrait of an old coal miner in 19th century, beautiful painting with highly detailed face by greg rutkowski and magali villanueve
```shell
一幅19世纪一位老煤矿工人的肖像，美丽的画作，面部非常细致
```
### 示例2:  
beautiful girl, ultra detailed eyes, mature, plump, thick, Opal drops, paint teardrops, woman made up from paint, entirely paint, splat, splash, long colored hair, kimono made from paint, ultra detailed texture kimono, Opalescent paint kimono, paint bulb
```shell
美丽的女孩，超细腻的眼睛，成熟，丰满，浓密，蛋白石滴，油漆泪珠，由油漆制成的女人，完全油漆，啪嗒，飞溅，长长的彩色头发，由油漆制成的和服，超精细纹理和服，乳白色油漆和服， 油漆灯泡
```
### 示例3：  
hyperrealistic portrait of gorgeous female tank commander, detailed gorgeous face, symmetric, intricate, realistic, hyperrealistic, cinematic
```shell
华丽的女坦克指挥官的超现实主义肖像，细致华丽的脸，对称，错综复杂，逼真，超现实主义，电影
```

服务端代码：
```python
# stable-diffusion: https://github.com/CompVis/stable-diffusion
from torch import autocast
from diffusers import StableDiffusionPipeline
import sys

pipe = StableDiffusionPipeline.from_pretrained(

	"stable-diffusion-v1-4",
	use_auth_token=True

).to("cuda")

prompt = "a photo of an astronaut riding a horse on mars"
prompt = sys.argv[1]
print("values: {} {}".format(sys.argv[1], sys.argv[2]))

with autocast("cuda"):
	image = pipe(prompt, num_inference_steps=100).images[0]
	image.save(sys.argv[2] + ".png")
```