import React from "react";
import { ThemeProvider } from "@material-ui/core";
import { createTheme } from "@material-ui/core";
import ExplorerView from "./views/ExplorerView";
import { BrowserRouter as Router, Route, Routes } from "react-router-dom";

if (process.env.NODE_ENV !== "production") {
  const whyDidYouRender = require("@welldone-software/why-did-you-render");
  whyDidYouRender(React);
}

const App: React.FC = () => {
  const darkTheme = createTheme({
    palette: {
      type: "dark",
    },
  });

  return (
    <ThemeProvider theme={darkTheme}>
      <Router>
        <Routes>
          <Route path="" element={<ExplorerView />} />
        </Routes>
      </Router>
    </ThemeProvider>
  );
};

export default App;
