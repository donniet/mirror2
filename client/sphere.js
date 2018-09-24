function createShader(gl, str, type) {
  var shader = gl.createShader(type);
  gl.shaderSource(shader, str);
  gl.compileShader(shader);
  if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
    throw new Error(gl.getShaderInfoLog(shader));
  }
  return shader;
}
function createProgram(gl, vertex_shader, fragment_shader) {
  var program = gl.createProgram();
  var vshader = createShader(gl, vertex_shader, gl.VERTEX_SHADER);
  var fshader = createShader(gl, fragment_shader, gl.FRAGMENT_SHADER);
  gl.attachShader(program, vshader);
  gl.attachShader(program, fshader);
  gl.linkProgram(program);
  if (!gl.getProgramParameter(program, gl.LINK_STATUS)) {
    throw new Error(gl.getProgramInfoLog(program));
  }
  return program;
}
function initTexturesAsync(gl, urls, nonPowerOfTwo) {
  function handleTextureLoad(gl, img, texture) {
    gl.bindTexture(gl.TEXTURE_2D, texture);
    gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, img);
    if (!nonPowerOfTwo) {
      gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
      gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_NEAREST);
      gl.generateMipmap(gl.TEXTURE_2D);
    } else {
      // gl.NEAREST is also allowed, instead of gl.LINEAR, as neither mipmap.
      gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
      // Prevents s-coordinate wrapping (repeating).
      gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
      // Prevents t-coordinate wrapping (repeating).
      gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
    }
    gl.bindTexture(gl.TEXTURE_2D, null);
  };

  var createTextureAsync = function (url) {
    return new Promise(function (resolve, reject) {
      var tex = gl.createTexture();

      var img;
      if (url instanceof Image || url instanceof HTMLImageElement) {
        handleTextureLoad(gl, url, tex);
        resolve(tex);
        return;
      } else if (url instanceof HTMLCanvasElement) {
        handleTextureLoad(gl, url, tex);
        resolve(tex);
        return;
      } else {
        img = new Image();
        img.onload = function () {
          handleTextureLoad(gl, img, tex);
          resolve(tex);
        };
        img.src = url;
      }
    });
  };
  if (typeof urls.map == 'function') {
    return Promise.all(urls.map(createTextureAsync));
  } else {
    return createTextureAsync(urls);
  }
}


function createSphereAsync(gl, texture_url) {
  "use strict";

  var corners = [-1, -1, 1, -1, 1, 1, -1, 1];

  var vertex = "\nprecision mediump float;\n\nattribute vec2 corner;\n\nuniform mat4 inv;\n\nvarying vec3 direction;\n\nvoid main() {\n  gl_Position = vec4(corner, 1.0, 1.0);\n  direction = (inv * gl_Position).xyz;\n}\n";

  var fragment = "\nprecision mediump float;\n\nuniform vec3 camera;\nuniform sampler2D texture;\nuniform vec3 sun;\nuniform float radius2;\nuniform vec3 position;\n\nvarying vec3 direction;\n\nvec4 textureSphere(sampler2D texture, vec3 p) {\n  vec2 s = vec2(atan(p.x, p.z), -asin(p.y));\n  /* [ \frac{1}{2pi}, \frac{1}{pi} ] */\n  s *= vec2(0.1591549430919, 0.31830988618379);\n\n  return texture2D(texture, vec2(s.s + 0.5, s.t + 0.5));\n}\n\nvoid swap(inout float a, inout float b) {\n  float c = a;\n  a = b;\n  b = c;\n}\n\nbool solveQuadratic(float a, float b, float c, out float x0, out float x1) {\n  float discr = b * b - 4. * a * c;\n  if (discr < 0.) return false;\n\n  if (discr == 0.) {\n    x0 = x1 = -0.5 * b / a;\n  } else {\n    float q;\n    if (b > 0.) {\n      q = -0.5 * (b + sqrt(discr));\n    } else {\n      q = -0.5 * (b - sqrt(discr));\n    }\n    x0 = q / a;\n    x1 = c / q;\n\n    if (x0 > x1) swap(x0, x1);\n  }\n\n  return true;\n}\n\n// https://www.scratchapixel.com/lessons/3d-basic-rendering/minimal-ray-tracer-rendering-simple-shapes/ray-sphere-intersection\nbool rayIntersectsSphere(vec3 orig, vec3 dir, vec3 center, float r2, out vec3 inter0, out vec3 n0, out vec3 inter1, out vec3 n1) {\n  vec3 L = orig - center;\n  float a = dot(dir, dir);\n  float b = 2. * dot(dir, L);\n  float c = dot(L, L) - r2;\n\n  float t0, t1;\n  if (!solveQuadratic(a, b, c, t0, t1)) {\n    return false;\n  }\n\n  if (t0 < 0.) {\n    t0 = t1;\n    if (t0 < 0.) return false;\n  }\n\n  inter0 = orig + t0 * dir;\n  n0 = normalize(inter0 - center);\n\n  inter1 = orig + t1 * dir;\n  n1 = normalize(inter1 - center);\n\n  return true;\n}\n\nvoid main() {\n  vec3 inter0, n0, inter1, n1;\n  if (rayIntersectsSphere(camera, direction, position, radius2, inter0, n0, inter1, n1)) {\n    float shade = max(dot(sun, n0), 0.) + 0.05;\n    vec4 col = textureSphere(texture, n0);\n    vec3 nd = n0 - n1;\n    float edge = dot(nd, nd) / 4. / radius2;\n    edge = 1. - pow(1. - edge, 100.);\n\n    gl_FragColor = vec4(edge * shade * col.xyz, col.w);\n  } else {\n    gl_FragColor = vec4(0.,0.,0.,1.);\n  }\n\n}\n";

  let Sphere = function () {
    function Sphere(gl, obj = {}) {
      this.texture = obj.texture;
      this.program = createProgram(gl, vertex, fragment);

      this.locations = {
        corner: gl.getAttribLocation(this.program, 'corner'),
        inv: gl.getUniformLocation(this.program, 'inv'),
        camera: gl.getUniformLocation(this.program, 'camera'),
        texture: gl.getUniformLocation(this.program, 'texture'),
        sun: gl.getUniformLocation(this.program, 'sun'),
        radius2: gl.getUniformLocation(this.program, 'radius2'),
        position: gl.getUniformLocation(this.program, 'position')
      };

      this.buffers = {};
      this.buffers.corners = gl.createBuffer();
      gl.bindBuffer(gl.ARRAY_BUFFER, this.buffers.corners);
      gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(corners), gl.STATIC_DRAW);
    }

    Sphere.prototype.render = function(gl, sun, model, view, projection) {
      var t = new Date().getTime();

      if (!model) model = mat4.create();
      if (!view) view = mat4.create();
      if (!projection) projection = mat4.create();
      var mv = mat4.create();
      mat4.multiply(mv, view, model);
      mat4.multiply(mv, projection, mv);
      mat4.invert(mv, mv);

      var invmodel = mat4.create();
      mat4.invert(invmodel, model);
      var position = vec4.clone([0, 0, 0, 1]);
      vec4.transformMat4(position, position, invmodel);

      var camera = vec4.clone([0, 0, 0, 1]);
      var iv = mat4.create();
      mat4.invert(iv, view);
      vec4.transformMat4(camera, camera, iv);

      if (!this.output) {
        this.output = true;
        console.log('camera', camera);
      }

      gl.useProgram(this.program);

      gl.enableVertexAttribArray(this.locations.corner);

      gl.uniformMatrix4fv(this.locations.inv, false, mv);
      gl.uniform3fv(this.locations.camera, camera.slice(0, 3));
      gl.uniform3fv(this.locations.sun, sun);
      gl.uniform1f(this.locations.radius2, 1);
      gl.uniform3fv(this.locations.position, position.slice(0, 3));

      gl.activeTexture(gl.TEXTURE0);
      gl.bindTexture(gl.TEXTURE_2D, this.texture);
      gl.uniform1i(this.locations.texture, 0);

      gl.bindBuffer(gl.ARRAY_BUFFER, this.buffers.corners);
      gl.vertexAttribPointer(this.locations.corner, 2, gl.FLOAT, false, 0, 0);
      gl.drawArrays(gl.TRIANGLE_FAN, 0, 4);

      gl.disableVertexAttribArray(this.locations.corner);
    }

    return Sphere;
  }();

  return Promise.all([initTexturesAsync(gl, texture_url, true)]).then(function ([moon]) {
    return new Sphere(gl, { texture: moon });
  });
}
