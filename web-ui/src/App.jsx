/* eslint-disable react-hooks/set-state-in-effect */

import { useEffect, useMemo, useState } from "react";
import { Container, Paper, Select, MenuItem, Typography } from "@mui/material";
// Anpassa path om din chart-komponent ligger annanstans
import PriceChart from "./components/PriceChart";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:18088";

const SYMBOLS = ["BTC-USD", "ETH-USD", "SOL-USD"];
const INTERVALS = ["all", "5m", "15m", "60m"];

export default function App() {
  const [symbol, setSymbol] = useState("BTC-USD");
  const [interval, setInterval] = useState("all");
  const [price, setPrice] = useState(null);
  const [history, setHistory] = useState([]);

  // "Nuvarande tid" som state
  const [now, setNow] = useState(() => Date.now());

  // Uppdatera "nu" var 4:e sekund så grafen ritas om
  useEffect(() => {
    const id = setInterval(() => {
      setNow(Date.now());
    }, 4000);

    return () => clearInterval(id);
  }, []);

  // --- Hämta senaste priset ---
  useEffect(() => {
    async function load() {
      try {
        const res = await fetch(`${API_URL}/prices/${symbol}`);
        const data = await res.json();
        if (!data.error) setPrice(data);
      } catch (err) {
        console.error("Failed to fetch latest price:", err);
      }
    }
    load();
  }, [symbol, now]); // uppdateras när symbol ändras eller tiden tickar

  // --- Hämta historik ---
  useEffect(() => {
    async function load() {
      try {
        const res = await fetch(
          `${API_URL}/prices/${symbol}/history?limit=200`
        );
        const data = await res.json();
        if (!data.error) setHistory(data.history);
      } catch (err) {
        console.error("Failed to fetch history:", err);
      }
    }
    load();
  }, [symbol]);

  // --- Filtrera historiken beroende på intervall ---
  const chartData = useMemo(() => {
    return history.filter((p) => {
      if (interval === "all") return true;

      const minutes =
        interval === "5m"
          ? 5
          : interval === "15m"
          ? 15
          : interval === "60m"
          ? 60
          : 9999;

      return now - new Date(p.timestamp).getTime() <= minutes * 60 * 1000;
    });
  }, [history, interval, now]);

  return (
    <Container maxWidth="md" sx={{ mt: 4 }}>
      <Paper sx={{ p: 3 }}>
        <Typography variant="h4" sx={{ mb: 2 }}>
          Realtime Crypto Prices
        </Typography>

        {/* Välj symbol */}
        <Select
          value={symbol}
          onChange={(e) => setSymbol(e.target.value)}
          sx={{ mr: 2 }}
        >
          {SYMBOLS.map((s) => (
            <MenuItem key={s} value={s}>
              {s}
            </MenuItem>
          ))}
        </Select>

        {/* Välj intervall */}
        <Select value={interval} onChange={(e) => setInterval(e.target.value)}>
          {INTERVALS.map((i) => (
            <MenuItem key={i} value={i}>
              {i}
            </MenuItem>
          ))}
        </Select>

        {/* Senaste priset */}
        {price && (
          <Typography sx={{ mt: 2 }}>
            Latest: <b>{price.price.toFixed(2)}</b> (vol {price.volume})
          </Typography>
        )}

        {/* Grafen */}
        <PriceChart data={chartData} />
      </Paper>
    </Container>
  );
}
