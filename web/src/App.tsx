import React from 'react';
import { createMuiTheme, ThemeProvider } from '@material-ui/core';
import ExplorerView from './views/ExplorerView';

if (process.env.NODE_ENV !== 'production') {
  const whyDidYouRender = require('@welldone-software/why-did-you-render');
  whyDidYouRender(React);
}

const App: React.FC = () => {
  const darkTheme = createMuiTheme({
    palette: {
      type: 'dark',
    },
  });

  return (
    <ThemeProvider theme={darkTheme}>
      <ExplorerView />
    </ThemeProvider >
  );
}

export default App;