import React from 'react';
import '../css/App.css';
//import NumbersContainer from './NumbersContainer';
import Navigation from './Nav';
import HistoryTableContainer from './HistoryTableContainer';
import NumbersContainer from './NumbersContainer';

function App() {
    return (
    <div className="App">
        <Navigation />
        <NumbersContainer />
    </div>
    );
    // <HistoryTableContainer />
    // <NumbersContainer />
}


export default App;
