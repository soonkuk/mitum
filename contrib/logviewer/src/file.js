import React, {useCallback} from 'react'
import {useDropzone} from 'react-dropzone'

function Dropzone() {
  const onDrop = useCallback(acceptedFiles => {
    /*
    console.log(acceptedFiles);
    const reader = new FileReader();

    reader.onabort = () => console.log('file reading was aborted');
    reader.onerror = () => console.log('file reading has failed');
    reader.onload = () => {
      const binaryStr = reader.result
      console.log(binaryStr)

      var log = Log.load(binaryStr);
    }

    acceptedFiles.forEach(file => reader.readAsBinaryString(file))
    */
  }, []);

  const {getRootProps, getInputProps, isDragActive, ...other} = useDropzone({onDrop})
  console.log(getRootProps())
  console.log(getInputProps())
  console.log('>>>>>>>', other)

  return (
    <div {...getRootProps()}>
      <input {...getInputProps()} />
      {
        isDragActive ?
          <p>Drop the files here ...</p> :
          <p>Drag 'n' drop some files here, or click to select files</p>
      }
    </div>
  )
}


export default Dropzone;
