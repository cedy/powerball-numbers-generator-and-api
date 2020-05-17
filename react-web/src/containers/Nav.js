import React from 'react';
import logo from '../logo.svg';
import Navbar from 'react-bootstrap/Navbar';
import Nav from 'react-bootstrap/Nav';

function Navigation() {
    return (
        <Navbar className="justify-content-center" expanded={false}>
            <Navbar.Brand className="mr-0" href="/home">
            <img className="App-logo" src={logo} alt="logo" />
            </Navbar.Brand>
                <Nav className="menu-text">
                    <Nav.Link className="mr-2" href="/home">Home</Nav.Link>
                    <Nav.Link className="mr-2" href="/history">History</Nav.Link>    
                    <Nav.Link href="/about">About</Nav.Link>    
                </Nav>
        </Navbar>
            )
}

export default Navigation;
