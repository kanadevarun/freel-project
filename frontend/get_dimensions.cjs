const fs = require('fs');
const sharp = require('sharp');

sharp('public/images/import-export/make-essential-documents.png')
  .metadata()
  .then(metadata => {
    console.log(metadata.width, metadata.height);
  });
