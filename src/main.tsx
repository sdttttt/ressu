import "semantic-ui-css/semantic.min.css";

import { render } from 'preact'
import { App } from './app'
import './index.css'

render(<App />, document.getElementById('app')!)
