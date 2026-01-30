import { Routes, Route } from "react-router-dom";
import { Layout } from "./components/Layout";
import { GatewayList } from "./components/GatewayList";
import { GatewayDetail } from "./components/GatewayDetail";

export function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<GatewayList />} />
        <Route path="gateways/:id" element={<GatewayDetail />} />
      </Route>
    </Routes>
  );
}
