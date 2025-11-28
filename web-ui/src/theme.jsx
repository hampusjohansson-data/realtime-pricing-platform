import { createTheme } from "@mui/material/styles";
import { useMemo } from "react";

export function useAppTheme(mode = "light") {
  return useMemo(() => {
    return createTheme({
      palette: {
        mode,
        ...(mode === "dark"
          ? {
              background: {
                default: "#121212",
                paper: "#1e1e1e",
              },
              text: {
                primary: "#ffffff",
                secondary: "#bbbbbb",
              },
            }
          : {
              background: {
                default: "#f6f6f6",
                paper: "#ffffff",
              },
              text: {
                primary: "#111111",
                secondary: "#444444",
              },
            }),
      },
      typography: {
        fontFamily: "Inter, sans-serif",
      },
    });
  }, [mode]);
}
