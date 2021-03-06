function fileSelected() {
  console.log("fileselected started.")
  var file = document.getElementById('fileToUpload').files[0];
  if (file) {
    var fileSize = 0;
    if (file.size > 1024 * 1024)
      fileSize = (Math.round(file.size * 100 / (1024 * 1024)) / 100).toString() + 'MB';
    else
      fileSize = (Math.round(file.size * 100 / 1024) / 100).toString() + 'KB';

    document.getElementById('fileName').innerHTML = 'Name: ' + file.name;
    document.getElementById('fileSize').innerHTML = 'Size: ' + fileSize;
    document.getElementById('fileType').innerHTML = 'Type: ' + file.type;
  }
}

function uploadFile(datasetURI) {
  var fd = new FormData();
  fd.append("fileToUpload", document.getElementById('fileToUpload').files[0]);
  var xhr = new XMLHttpRequest();
  xhr.upload.addEventListener("progress", uploadProgress, false);
  xhr.addEventListener("load", uploadComplete, false);
  xhr.addEventListener("error", uploadFailed, false);
  xhr.addEventListener("abort", uploadCanceled, false);
  xhr.open("POST", "/upload/"+datasetURI);
  xhr.send(fd);
}

function uploadProgress(evt) {
  if (evt.lengthComputable) {
    var percentComplete = Math.round(evt.loaded * 100 / evt.total);
    //document.getElementById('progressNumber').innerHTML = percentComplete.toString() + '%';
    //$(".uploadProgressBar").css("width", percentComplete.toString() + '%');
    //var uploadBar = document.getElementById("uploadProgressBar");
    //uploadBar.style.width = percentComplete.toString() + '%';
    console.log("progress: " + percentComplete.toString());
    $('#uploadProgressBar.progress-bar').css('width', percentComplete.toString()+'%').attr('aria-valuenow', percentComplete.toString()); 
  }
  else {
    document.getElementById('flash').innerHTML = 'unable to compute';
  }
}

function uploadComplete(evt) {
  /* This event is raised when the server send back a response */
  // alert(evt.target.responseText);
  window.location = evt.target.responseText
}

function uploadFailed(evt) {
  document.getElementById('flash').innerHTML = 'Failed to upload'
    + ' the file';
}

function uploadCanceled(evt) {
  document.getElementById('flash').innerHTML = 'Upload canceled';
}

function triggerUpload() {
  var fd = new FormData();
  var createXHR = new XMLHttpRequest();
  createXHR.addEventListener("load", createComplete, false);
  createXHR.addEventListener("error", createFailed, false);
  createXHR.addEventListener("abort", createFailed, false);
  createXHR.open("POST", "/createcollection");
  createXHR.send(fd);
}

function createComplete(evt) {
  var response = JSON.parse(evt.target.responseText)
    document.getElementById('flash').innerHTML = 'Created dataset '
    + response.URI
    uploadFile(response.URI);
}

function createFailed() {
  document.getElementById('flash').innerHTML = 'Failed to create'
    + ' dataset - sorry!';
}

