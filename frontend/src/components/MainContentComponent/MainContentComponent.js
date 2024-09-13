import React, {useState, useEffect} from 'react';
import { Box, Container, Alert, Grid, AlertTitle, Table, TableContainer, TableHead, TableCell, TableBody, TableRow, Paper } from '@mui/material';
import axios from 'axios';

const apiURL = process.env.REACT_APP_BACKEND_URL;

const MainContentComponent = () => {

    const [files, setFiles] = useState([]);
    const [dataCourses, setDataCourses] = useState(0);
    const [dataStudentRegistration, setDataStudentRegistration] = useState(0);
    const [dataAssessments, setDataAssessments] = useState(0);
    const [dataVle, setDataVle] = useState(0);
    const [dataStudentInfo, setDataStudentInfo] = useState(0);
    const [dataStudentAssessment, setDataStudentAssessment] = useState(0);
    const [dataStudentVle, setDataStudentVle] = useState(0);
    const [dataStudentAssessmentPrediction, setDataStudentAssessmentPrediction] = useState([]);
    const [isExistingData, setIsExistingData] = useState(false);

    useEffect(() => {
        const fetchFiles = async () => {
                try {
                    const response = await axios.get(`${apiURL}/get_files`);
                    let data = response.data;
                    if (!data) {
                        console.error('No data found');
                        return;
                    }
                    data = data.filter((file) => file.file_name.includes('.csv'));
                    console.log('Files:', data);
                    setFiles(data);
                } catch (error) {
                    console.error('Error fetching files:', error);
                }
            };
            fetchFiles();

            const collections = JSON.parse(process.env.REACT_APP_COLLECTIONS);

            const fetchData = async () => {
                try {
                    const response = await axios.post(`${apiURL}/get_all_data`, { collections });
                    console.log('Collections processed:', response.data);
                    setDataCourses(response.data.courses);
                    setDataStudentRegistration(response.data.studentRegistration);
                    setDataAssessments(response.data.assessments);
                    setDataVle(response.data.vle);
                    setDataStudentInfo(response.data.studentInfo);
                    setDataStudentAssessment(response.data.studentAssessment);
                    setDataStudentVle(response.data.studentVle)

                } catch (error) {
                    console.error('Error processing collections:', error);
                }
            };
            fetchData();

            const fetchPredictionAssessments = async () => {
                try {
                    const fetchPredictions = await axios.get(`${apiURL}/get_data/predictions_assestments`);
                    console.log('Predictions:', fetchPredictions.data);
                    setDataStudentAssessmentPrediction(fetchPredictions.data);
                } catch (error) {
                    console.error('Error fetching prediction data:', error);
                }
            };
            fetchPredictionAssessments();

            console.log('Data:', dataCourses, dataStudentRegistration, dataAssessments, dataVle, dataStudentInfo, dataStudentAssessment, dataStudentVle);
            
    }, []);
    


    return (
        <Container component="main" maxWidth="md" sx={{ mt: 10, p: 3, bgcolor: 'background.paper', boxShadow: 3, borderRadius: 2, minWidth:'80%', minHeight: '100%' }}>
            <Grid container spacing={1}>
                <Grid item container xs={4}>
                    {files.length > 0 ? (
                        <Alert severity="info" sx={{ minHeight: 200, minWidth:'100%' }}>
                            <AlertTitle>{files.length} files found</AlertTitle>
                            <Grid container spacing={2}>
                                <TableContainer component={Box}>
                                    <Table>
                                        <TableBody>
                                            {files.map((file, index) => (
                                                <TableRow key={index}>
                                                    <TableCell>{file.file_name}</TableCell>
                                                    <TableCell><b>{file.file_size}</b></TableCell>
                                                </TableRow>
                                            ))}
                                        </TableBody>
                                    </Table>
                                </TableContainer>
                            </Grid>
                        </Alert>
                    ) : (
                        <Alert severity="error" sx={{ minHeight: 200, minWidth:'100%' }}>
                            No files found
                        </Alert>
                    )}
                </Grid>
                <Grid item container xs={4}>
                        <Alert severity="info" sx={{ minHeight: 200, minWidth:'100%' }}>
                            <AlertTitle>Documents Inserted</AlertTitle>
                            <Grid container spacing={2}>
                                <TableContainer component={Box}>
                                    <Table>
                                        <TableBody>
                                            <TableRow>
                                                <TableCell>Assessments: <b>{new Intl.NumberFormat('en-US').format(dataAssessments)}</b></TableCell>
                                            </TableRow>    
                                            <TableRow>
                                                <TableCell>Courses: <b>{new Intl.NumberFormat('en-US').format(dataCourses)}</b></TableCell>
                                            </TableRow>    
                                            <TableRow>    
                                                <TableCell>Student-Assessments: <b>{new Intl.NumberFormat('en-US').format(dataStudentAssessment)}</b></TableCell>
                                            </TableRow>    
                                            <TableRow>    
                                                <TableCell>Student-Info: <b>{new Intl.NumberFormat('en-US').format(dataStudentInfo)}</b></TableCell>
                                            </TableRow>    
                                            <TableRow>    
                                                <TableCell>Student-Registration: <b>{new Intl.NumberFormat('en-US').format(dataStudentRegistration)}</b></TableCell>
                                            </TableRow>    
                                            <TableRow>    
                                                <TableCell>Student-VLE: <b>{new Intl.NumberFormat('en-US').format(dataStudentVle)}</b></TableCell>
                                            </TableRow>    
                                            <TableRow>    
                                                <TableCell>VLE: <b>{new Intl.NumberFormat('en-US').format(dataVle)}</b></TableCell>
                                            </TableRow>
                                        </TableBody>
                                    </Table>
                                </TableContainer>
                            </Grid>
                        </Alert>
                </Grid>
                <Grid item container xs={4}>
                    {isExistingData ? (
                        <Alert severity="info" sx={{ minHeight: 200, minWidth:'100%' }}>
                            <AlertTitle>Documents Inserted</AlertTitle>
                            <Grid container spacing={2}>
                                <TableContainer component={Box}>
                                    <Table>
                                        <TableBody>
                                            {files.map((file, index) => (
                                                <TableRow key={index}>
                                                    <TableCell key={index}>{file.file_name}</TableCell>
                                                </TableRow>
                                            ))}
                                        </TableBody>
                                    </Table>
                                </TableContainer>
                            </Grid>
                        </Alert>
                    ) : (
                        <Alert severity="error" sx={{ minHeight: 200, minWidth:'100%' }}>
                            No files found
                        </Alert>
                    )}
                </Grid>
            </Grid>
            
            
            
        </Container>
        
    );
};

export default MainContentComponent;
