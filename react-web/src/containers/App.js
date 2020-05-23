import React from 'react';
import '../css/App.css';
//import NumbersContainer from './NumbersContainer';
import Navigation from './Nav';
import HistoryTableContainer from './HistoryTableContainer';
import NumbersContainer from './NumbersContainer';
import About from './About';
import {
    BrowserRouter as Router,
    Switch,
    Route
} from 'react-router-dom';


function App() {
    return (
    <div className="App">
        <Router>
            <Navigation />
            <Switch>
                <Route path="/history">
                    <HistoryTableContainer />
                </Route>
                <Route path="/numbers">
                    <NumbersContainer />
                </Route>
                <Route path="/about">
                    <About />
                </Route>
            </Switch>
        </Router>
    </div>

    );
    // <HistoryTableContainer />
    // <NumbersContainer />
}


export default App;
