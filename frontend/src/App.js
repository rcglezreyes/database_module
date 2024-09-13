
import './App.css';
import LayoutComponent from './components/LayoutComponent/LayoutComponent';
import { BrowserRouter as Router, Route, Sw, Routes } from 'react-router-dom';
import MainContentComponent from './components/MainContentComponent/MainContentComponent';
import InfoComponent from './components/InfoComponent/InfoComponent';

function App() {
  return (
    <Router>
        <Routes>
            <Route path="/*" element={<LayoutComponent/>}>
              <Route path="" element={<MainContentComponent/>} />
              <Route path="info/download" element={<InfoComponent/>} />
              <Route path="info/upload" element={<InfoComponent/>} />
              <Route path="info/process_prediction_assessments" element={<InfoComponent/>} />
            </Route>
        </Routes>
    </Router>
  );
}

export default App;
