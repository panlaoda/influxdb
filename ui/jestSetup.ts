import {cleanup} from 'react-testing-library'
import 'intersection-observer'

process.env.API_PREFIX = '/'
// cleans up state between react-testing-library tests
afterEach(() => {
  cleanup()
})
