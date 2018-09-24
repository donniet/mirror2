
function Reloader() {
  this.lastModified = '';
  this.checker = null;

  this.checkHeader();
}

Reloader.prototype.checkHeader = function() {
  getLastModifiedDateAsync(window.location).then(lastModified => {
    if (this.lastModified != '' && this.lastModified != lastModified) {
      window.location.reload();
      return;
    }

    this.lastModified = lastModified;
    this.checker = setTimeout(() => this.checkHeader(), 1000);
  })
}
