$(document).ready ( function(){

  function startUpload(r, file) {
    // Show progress pabr
    $('.resumable-progress, .resumable-list').show();
    // Show pause, hide resume
    $('.resumable-progress .progress-resume-link').hide();
    $('.resumable-progress .progress-pause-link').show();
    // Add the file to the list
    $('.resumable-list').append('<li class="resumable-file-'+file.uniqueIdentifier+'">Uploading <span class="resumable-file-name"></span> <span class="resumable-file-progress"></span>');
    $('.resumable-file-'+file.uniqueIdentifier+' .resumable-file-name').html(file.fileName);
    // Actually start the upload
    console.log("Uploading to " + r.opts['target']);
    r.upload();
  }

  var r = new Resumable({
      target:'',
      chunkSize:1*1024*1024,
      simultaneousUploads:4,
      testChunks:false,
      throttleProgressCallbacks:1
  });
  // Resumable.js isn't supported, fall back on a different method
  if(!r.support) {
    $('.resumable-error').show();
    $('.resumable-drop').hide();
  } else {
    $('.resumable-error').hide();
    // Show a place for dropping/selecting files
    $('.resumable-drop').show();
    r.assignDrop($('.resumable-drop')[0]);
    r.assignBrowse($('.resumable-browse')[0]);

    // Handle file add event
    r.on('fileAdded', function(file){
      if (r.opts['target'] === '') {
        console.log('Attempting to create collection');
        var fd = new FormData();
        var createXHR = new XMLHttpRequest();
        createXHR.addEventListener("load", function(evt) {
          var response = JSON.parse(evt.target.responseText)
          document.getElementById('flash').innerHTML = 'Created dataset '
            + response.URI
          // We have created our collection. Now: Start the upload
          r.opts['target']="/upload/" + response.URI;
          r.opts['target_collection']= response.URI;
          startUpload(r, file)
        } , false);
        createXHR.addEventListener("error", function(evt) {
          document.getElementById('flash').innerHTML = 'Error while creating'
            + ' dataset - sorry!';
        }, false);
        createXHR.addEventListener("abort", function(evt) {
          document.getElementById('flash').innerHTML = 'Aborted creating'
            + ' dataset - sorry!';
        }, false);
        createXHR.open("POST", "/createcollection");
        createXHR.send(fd);
      } else {
        // we already have a collection.
        startUpload(r, file);
      }
    });

    r.on('pause', function(){
      // Show resume, hide pause
      $('.resumable-progress .progress-resume-link').show();
      $('.resumable-progress .progress-pause-link').hide();
    });

    r.on('complete', function(){
      // Hide pause/resume when the upload has completed
      $('.resumable-progress .progress-resume-link, .resumable-progress .progress-pause-link').hide();
      // redirect to collection page
      window.location = "/show/" + r.opts['target_collection'];
    });

    r.on('fileSuccess', function(file,message){
      // Reflect that the file upload has completed
      $('.resumable-file-'+file.uniqueIdentifier+' .resumable-file-progress').html('(completed)');
    });

    r.on('fileError', function(file, message){
      // Reflect that the file upload has resulted in error
      $('.resumable-file-'+file.uniqueIdentifier+' .resumable-file-progress').html('(file could not be uploaded: '+message+')');
    });

    r.on('fileProgress', function(file){
      // Handle progress for both the file and the overall upload
      $('.resumable-file-'+file.uniqueIdentifier+' .resumable-file-progress').html(Math.floor(file.progress()*100) + '%');
      $('.progress-bar').css({width:Math.floor(r.progress()*100) + '%'});
    });

    r.on('cancel', function(){
      $('.resumable-file-progress').html('canceled');
    });

    r.on('uploadStart', function(){
      // Show pause, hide resume
      $('.resumable-progress .progress-resume-link').hide();
      $('.resumable-progress .progress-pause-link').show();
    });
  }
});
