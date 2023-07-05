import React, {useState} from 'react'
import ReactDOM from 'react-dom/client'
import {
  RouterProvider,
  Router,
  Route,
  RootRoute,
} from '@tanstack/router'
import {ColorScheme, ColorSchemeProvider, MantineProvider} from "@mantine/core";

import App, {HomePage} from './App.tsx'

// TODO: Pull routing into seperate file!
const rootRoute = new RootRoute({
  component: App,
})

const indexRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/',
  component: HomePage,
})

const routeTree = rootRoute.addChildren([indexRoute])

const router = new Router({ routeTree })

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
        <RouterProvider router={router} />
      </MantineProvider>
    </ColorSchemeProvider>
  </React.StrictMode>
}

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <Root/>
)
