import React, { useState } from "react";
import ReactDOM from "react-dom/client";
import {
  ColorScheme,
  ColorSchemeProvider,
  MantineProvider,
} from "@mantine/core";

import { ApprovalQueuePage, HomePage, LoginPage, Shell } from "./App.tsx";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { AgentProvider } from "./auth.tsx";

const router = createBrowserRouter([
  {
    path: "/",
    element: <Shell />,
    children: [
      {
        path: "/",
        element: <HomePage />,
      },
      {
        path: "/approval-queue",
        element: <ApprovalQueuePage />,
      },
    ],
  },
  {
    path: "/login",
    element: <LoginPage />,
  },
]);

const Root = () => {
  const [colorScheme, setColorScheme] = useState<ColorScheme>("light");
  const toggleColorScheme = (value?: ColorScheme) =>
    setColorScheme(value || (colorScheme === "dark" ? "light" : "dark"));

  return (
    <React.StrictMode>
      <AgentProvider>
        <ColorSchemeProvider
          colorScheme={colorScheme}
          toggleColorScheme={toggleColorScheme}
        >
          <MantineProvider withGlobalStyles withNormalizeCSS>
            <RouterProvider router={router} />
          </MantineProvider>
        </ColorSchemeProvider>
      </AgentProvider>
    </React.StrictMode>
  );
};

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <Root />,
);
