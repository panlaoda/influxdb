// Libraries
import React, {PureComponent} from 'react'
import {Controlled as ReactCodeMirror, IInstance} from 'react-codemirror2'
import {EditorChange, LineWidget, Position} from 'codemirror'
import {ShowHintOptions} from 'src/types/codemirror'
import 'src/external/codemirror'
import 'codemirror/addon/comment/comment'

// Components
import {ErrorHandling} from 'src/shared/decorators/errors'

// Constants
import {EXCLUDED_KEYS} from 'src/shared/constants/fluxEditor'

// Utils
import {getSuggestions} from 'src/shared/utils/autoComplete'
import {onTab} from 'src/shared/utils/fluxEditor'

// Types
import {OnChangeScript, Suggestion} from 'src/types/flux'

interface Gutter {
  line: number
  text: string
}

interface Status {
  type: string
  text: string
}

interface Props {
  script: string
  status?: Status
  onChangeScript: OnChangeScript
  onSubmitScript?: () => void
  suggestions: Suggestion[]
  visibility?: string
  onCursorChange?: (position: Position) => void
}

interface Widget extends LineWidget {
  node: HTMLElement
}

interface State {
  lineWidgets: Widget[]
}

interface EditorInstance extends IInstance {
  showHint: (options?: ShowHintOptions) => void
}

@ErrorHandling
class FluxEditor extends PureComponent<Props, State> {
  public static defaultProps = {
    visibility: 'visible',
    status: {text: '', type: ''},
  }

  private editor: EditorInstance
  private lineWidgets: Widget[] = []

  constructor(props) {
    super(props)
  }

  public componentDidUpdate(prevProps) {
    const {status, visibility} = this.props

    if (status.type === 'error') {
      this.makeError()
    }

    if (status.type !== 'error') {
      this.editor.clearGutter('error-gutter')
      this.clearWidgets()
    }

    if (prevProps.visibility === visibility) {
      return
    }

    if (visibility === 'visible') {
      setTimeout(() => this.editor.refresh(), 60)
    }
  }

  public render() {
    const {script} = this.props

    const options = {
      tabIndex: 1,
      mode: 'flux',
      readonly: false,
      lineNumbers: true,
      autoRefresh: true,
      theme: 'time-machine',
      completeSingle: false,
      gutters: ['error-gutter'],
      comment: true,
      extraKeys: {
        Tab: onTab,
        'Ctrl-/': 'toggleComment',
      },
    }

    return (
      <div className="time-machine-editor" data-testid="flux-editor">
        <ReactCodeMirror
          autoFocus={true}
          autoCursor={true}
          value={script}
          options={options}
          onBeforeChange={this.updateCode}
          onTouchStart={this.onTouchStart}
          editorDidMount={this.handleMount}
          onKeyUp={this.handleKeyUp}
          onCursor={this.handleCursorChange}
        />
      </div>
    )
  }

  private handleCursorChange = (__: IInstance, position: Position) => {
    const {onCursorChange} = this.props

    if (onCursorChange) {
      onCursorChange(position)
    }
  }

  private makeError(): void {
    this.editor.clearGutter('error-gutter')
    const lineNumbers = this.statusLine
    lineNumbers.forEach(({line, text}) => {
      const lineNumber = line - 1
      this.editor.setGutterMarker(
        lineNumber,
        'error-gutter',
        this.errorMarker(text, lineNumber)
      )
    })

    this.editor.refresh()
  }

  private errorMarker(message: string, line: number): HTMLElement {
    const span = document.createElement('span')
    span.className = 'icon stop error-warning'
    span.title = message
    span.addEventListener('click', this.handleClickError(message, line))
    return span
  }

  private handleClickError = (text: string, line: number) => () => {
    let widget = this.lineWidgets.find(w => w.node.textContent === text)

    if (widget) {
      return this.clearWidget(widget)
    }

    const errorDiv = document.createElement('div')
    errorDiv.className = 'inline-error-message'
    errorDiv.innerText = text
    widget = this.editor.addLineWidget(line, errorDiv) as Widget

    this.lineWidgets = [...this.lineWidgets, widget]
  }

  private clearWidget = (widget: Widget): void => {
    widget.clear()
    this.lineWidgets = this.lineWidgets.filter(
      w => w.node.textContent !== widget.node.textContent
    )
  }

  private clearWidgets = () => {
    this.lineWidgets.forEach(w => {
      w.clear()
    })

    this.lineWidgets = []
  }

  private get statusLine(): Gutter[] {
    const {status} = this.props
    const messages = status.text.split('\n')
    const lineNumbers = messages.map(text => {
      const [numbers] = text.split(' ')
      const [lineNumber] = numbers.split(':')
      return {line: Number(lineNumber), text}
    })

    return lineNumbers
  }

  private handleMount = (instance: EditorInstance) => {
    instance.refresh() // required to for proper line height on mount
    this.editor = instance
  }

  private onTouchStart = () => {}

  private handleKeyUp = (__, e: KeyboardEvent) => {
    const {ctrlKey, metaKey, key} = e
    const {onSubmitScript} = this.props

    if (ctrlKey && key === 'Enter' && onSubmitScript) {
      onSubmitScript()
    }

    if (ctrlKey && key === ' ') {
      this.showAutoComplete()

      return
    }

    if (ctrlKey || metaKey || EXCLUDED_KEYS.includes(key)) {
      return
    }

    this.showAutoComplete()
  }

  private showAutoComplete() {
    const {suggestions} = this.props

    this.editor.showHint({
      hint: () => getSuggestions(this.editor, suggestions),
      completeSingle: false,
    })
  }

  private updateCode = (
    _: IInstance,
    __: EditorChange,
    script: string
  ): void => {
    this.props.onChangeScript(script)
  }
}

export default FluxEditor
