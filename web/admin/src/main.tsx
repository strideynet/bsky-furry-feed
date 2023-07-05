import React, {useState} from 'react'
import ReactDOM from 'react-dom/client'
import App from './App.tsx'
import {ColorScheme, ColorSchemeProvider, MantineProvider} from "@mantine/core";

const Root = () => {
  const [colorScheme, setColorScheme] = useState<ColorScheme>('light');
  const toggleColorScheme = (value?: ColorScheme) =>
    setColorScheme(value || (colorScheme === 'dark' ? 'light' : 'dark'));

  return <React.StrictMode>
    <ColorSchemeProvider
      colorScheme={colorScheme}
      toggleColorScheme={toggleColorScheme}
    >
      <MantineProvider withGlobalStyles withNormalizeCSS>
        <App/>
      </MantineProvider>
    </ColorSchemeProvider>
  </React.StrictMode>
}

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <Root/>
)
