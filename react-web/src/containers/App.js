import React from 'react';
import '../css/App.css';
//import NumbersContainer from './NumbersContainer';
import Navigation from './Nav';
import HistoryTableContainer from './HistoryTableContainer';

function App() {
    return (
    <div className="App">
        <Navigation />
        <HistoryTableContainer />
    </div>
    );

    // <NumbersContainer />
}


export default App;
