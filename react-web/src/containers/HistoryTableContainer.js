import React from 'react';
import HistoryRow from '../components/HistoryRow';
import PPagination from '../components/Pagination';
import Table from 'react-bootstrap/Table';

const API_ADDRESS = "http://127.0.0.1:8080"

class HistoryTableContainer extends React.Component {
    constructor(props) {
        super(props);
        this.onClickFirstPage = this.onClickFirstPage.bind(this);
        this.onClickNextPage = this.onClickNextPage.bind(this);
        this.onClickPreviousPage = this.onClickPreviousPage.bind(this);
        this.state = {
            rows: {},
            currentPage: 1,
        }
    }
    
    componentDidMount() {
        this.loadPageData(this.state.currentPage);
    }

    errorHadnler(error) {
                    alert("Something went wrong, please refresh the page");
                    console.log(error);
    }
    
    loadPageData(pageNumber) {
        fetch(API_ADDRESS + "/history/page/" + pageNumber)
            .then((res) => res.json())
            .then(
                (result) => {
                    let rows = {};
                    result.forEach((element) => {
                        rows[element.numbers] = {
                            date: element.date, 
                            rCount: {day: 0, week: 0, month: 0, year: 0, allTime: 0}
                        };
                    });
                    this.setState({rows: rows});
                    return rows;
                })
            .then((rows) => {
                Object.keys(rows).forEach((key) => {
                    fetch(API_ADDRESS + "/numbers/" + key.split(" ").join("/"))
                            .then(res => res.json())
                            .then((res) => {
                                if (res.length  === 1) {
                                    rows[key].rCount = {
                                        day: res[0].dayCount, week: res[0].weekCount,
                                        month: res[0].weekCount, year: res[0].yearCount,
                                        allTime: res[0].count};
                                    this.setState({rows: rows});
                                }
                            });
                            });
            })
            .catch(this.errorHadnler);
    }

    onClickFirstPage(event) {
        if (this.state.currentPage === 1) {
            return
        }
        this.setState({currentPage: 1});
        this.loadPageData(1);
    }

    onClickNextPage(event) {
        if (Object.keys(this.state.rows).length < 100) {
            return
        }
        let currentPage = this.state.currentPage + 1 ;
        this.setState({currentPage: currentPage});
        this.loadPageData(currentPage);
    }

    onClickPreviousPage(event) {
        if (this.state.currentPage === 1) {
            return
        }
        let currentPage = this.state.currentPage - 1 ;
        this.setState({currentPage: currentPage});
        this.loadPageData(currentPage);
    }

    render() {
        const RGNC = <abbr title="Randomly Generated Numbers Count">RGNC</abbr>;
        let tableRows = []; 
        if (Object.keys(this.state.rows).length){
        Object.keys(this.state.rows).forEach((key) => {
            tableRows.push(
                <HistoryRow key={key} date={this.state.rows[key].date} numbers={key} dayCount={this.state.rows[key].rCount['day']} weekCount={this.state.rows[key].rCount.week} monthCount={this.state.rows[key].rCount.month} yearCount={this.state.rows[key].rCount.year} allTime={this.state.rows[key].rCount.allTime}/>
        );
        });
        };
        return (
            <div className="main-container">
            <Table striped bordered hover>
                <thead>
                    <tr className="text-center">
                        <th className="align-middle">Date</th>
                        <th className="align-middle">Numbers</th>
                        <th className="align-middle">{RGNC} Daily</th>
                        <th className="align-middle">{RGNC} Weekly</th>
                        <th className="align-middle">{RGNC} Monthly</th>
                        <th className="align-middle">{RGNC} Yearly</th>
                        <th className="align-middle">{RGNC} Total</th>
                    </tr>
                </thead>
                <tbody>
                    {tableRows}
                </tbody>

            </Table>
            <div className="d-flex justify-content-center">
                <PPagination
                    currentPage={this.state.currentPage}
                    first={this.onClickFirstPage}
                    previous={this.onClickPreviousPage}
                    next={this.onClickNextPage}
                />
            </div>
        </div>
        )
    }
}

export default HistoryTableContainer;
