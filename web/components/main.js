import styles from '../styles/Home.module.css'
import Head from 'next/head'
import React from 'react'
import Editor, { DiffEditor, useMonaco, loader } from "@monaco-editor/react";

export default function Main() {
  const [error, setError] = React.useState('');
  // This can be mutated when the editor is mounted. This
  // allows us to call run() on a page load.
  let [editor, setEditor] = React.useState(null);
  const [output, setOutput] = React.useState('');

  async function handleEditorDidMount(e, _) {
    setEditor(e);

    // If we have an ID, load the contents in
    const id = window.location.href.split('#')[1]
    if (!id) {
      return
    }

    const req = await fetch(`/api/v1/code/${id}`)
    const res = await req.json()
    e.setValue(res.contents)
    editor = e
    run()
  }

  async function run() {
    const code = editor.getValue();
    const req = await fetch('/api/v1/execute', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        code: code
      }),
    })
    const res = await req.json()
    if (res.message) {
      setError(res.message)
      setOutput('')
    }
    if (res.output) {
      setOutput(res.output)
      setError('')
    }
  }

  async function share() {
    const code = editor.getValue();
    const req = await fetch('/api/v1/code', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        code: code
      }),
    })
    const res = await req.json()
    window.location.href = `/#${res.id}`
  }


  return (
    <div>
      <Head>
        <title>Jsonnet Playground | jsonnet.dev</title>
        <meta name="title" content="Jsonnet Playground" />
        <meta name="description" content="Jsonnet playground with running, saving, and more" />
        <meta name="keywords" content="jsonnet,playground,editor,online" />
        <meta name="robots" content="index, follow" />
        <meta httpEquiv="Content-Type" content="text/html; charset=utf-8" />
        <meta name="language" content="English" />
      </Head>
      <div className="light tabwidth-4 ok withsidebar">
        <div className="header">
          <div className="logo"></div>
          <div className="menu">
            <span className="title">Jsonnet Playground</span>
            <button onClick={run}>Run</button>
            <button onClick={share}>Share</button>
          </div>
          <div className="settings">
            <a href="https://github.com/rgst-io/jsonnet-playground" className="button">Github</a>
          </div>
        </div>
        <div className="body-wrapper">
          <div className="content-wrapper">
            <div className="editor-wrapper">
              <Editor
                height="90vh"
                defaultLanguage="text"
                defaultValue="{}"
                onMount={handleEditorDidMount}
              />
              <div className="shadow"><ol></ol></div>
            </div>
            <div className="output-wrapper">
              <div className="help">
                <pre>{output}</pre>
              </div>
            </div>
          </div>
        </div>
        <div className="splitter col"></div>
        <div className="log-wrapper">
          <div className="log">
            <div></div>
            <div className={`status ${error ? 'error' : ''}`}>{error}</div>
            <div className="splitter row"></div>
          </div>
        </div>
      </div>
    </div>
  )
}
