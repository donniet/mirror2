
function getXPathText(node, xpath) {
  var res = node.ownerDocument.evaluate(xpath, node, null, XPathResult.ANY, null);
  var ret = "";
  var item = res.iterateNext();
  while(item) {
    ret += item.nodeValue;
    item = res.iterateNext();
  }
  return ret;
}

function getNewsAsync() {
  return ;
}

function clearNode(node) {
  while(node.firstChild) node.removeChild(node.firstChild);
}

function headAsync(url) {
  return new Promise(function (resolve, reject) {
    var xhr = new XMLHttpRequest();
    xhr.timeout = 2000;
    xhr.onreadystatechange = function(e) {
      switch(this.readyState) {
      case this.HEADERS_RECEIVED:
        var headers = request.getAllResponseHeaders();
        console.log(this.status, headers);
        xhr.onreadystatechange = null;
        resolve(headers);
        break;
      case this.DONE:
        console.log(this.status);
        resolve(null);
        break;
      }
    };
    xhr.open('HEAD', url, true);
    xhr.send();
  });
}

function checkImageAsync(url) {
  return new Promise(function (resolve, reject) {
    console.log('inside check image promise', url);
    var img = new Image();
    img.onload = () => { resolve(true); };
    img.onerror = () => { resolve(false); };
    img.src = url;
  });
}

function getXMLAsync(url) {
  return new Promise(function (resolve, reject) {
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function(e) {
      if(this.readyState == this.DONE) {
        if(this.status == 200) {
          resolve(xhr.responseXML);
        } else {
          resolve('error: ' + this.status);
        }
      }
    }
    xhr.open('GET', url, true);
    xhr.send();
  });
}

function getTextAsync(url) {
  return new Promise(function (resolve, reject) {
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function(e) {
      if(this.readyState == this.DONE) {
        if(this.status == 200) {
          resolve(xhr.responseText);
        } else {
          resolve('error: ' + this.status);
        }
      }
    }
    xhr.open('GET', url, true);
    xhr.send();
  });
}

function postAsync(url, body) {
  return new Promise(function(resolve, reject) {
    var xhr = new XMLHttpRequest();
    xhr.responseType = 'json';
    xhr.onreadystatechange = function(e) {
      if (this.readyState == this.DONE) {
        if (this.status == 200) {
          resolve(true);
        } else {
          resolve(false, 'error: ' + this.status);
        }
      }
    }
    if (typeof body !== "string") {
      body = JSON.stringify(body);
    }
    xhr.open('POST', url, true);
    xhr.send(body);
  });
}

function getJSONAsync(url) {
  return new Promise(function(resolve,reject) {
    var xhr = new XMLHttpRequest();
    xhr.responseType = 'json';
    xhr.onreadystatechange = function(e) {
      if(this.readyState == this.DONE) {
        if(this.status == 200) {
          resolve(xhr.response);
        } else {
          resolve(null, 'error: ' + this.status);
        }
      }
    }
    xhr.open('GET', url, true);
    xhr.send();
  });
}

function getLastModifiedDateAsync(url) {
  return new Promise(function (resolve, reject) {
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function(e) {
      if(this.readyState == this.HEADERS_RECEIVED) {
        if(this.status == 200) {
          resolve(this.getResponseHeader('Last-Modified'));
        }
      }
    }
    xhr.open('HEAD', url, true);
    xhr.send();
  });
}
