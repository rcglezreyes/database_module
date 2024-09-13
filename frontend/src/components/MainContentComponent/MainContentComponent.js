import React, { useState, useEffect } from 'react';
import { Box, Container, Alert, Grid, AlertTitle, Table, TableContainer, TableHead, TableCell, TableBody, TableRow, Paper, Typography } from '@mui/material';
import axios from 'axios';
import ScorePieChartComponent from '../ScorePieChartComponent/ScorePieChartComponent';
import AssessmentByIdBarChartComponent from '../AssessmentByIdBarChartComponent/AssessmentByIdBarChartComponent';

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
    const [dataStudentAssessmentPrediction, setDataStudentAssessmentPrediction] = useState(null);
    const [dataCountByAssessmentID, setDataCountByAssessmentID] = useState(null);
    const [dataAverageType, setDataAverageType] = useState(null);
    const [isExistingData, setIsExistingData] = useState(false);
    const [isExistingDataCountByAssessmentID, setIsExistingDataCountByAssessmentID] = useState(false);
    const [isExistingDataAverageType, setIsExistingDataAverageType] = useState(false);

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
                const fetchPredictions = await axios.get(`${apiURL}/get_score_distribution_prediction_assessments`);
                console.log('Predictions:', fetchPredictions.data);
                setDataStudentAssessmentPrediction(fetchPredictions.data);
                setIsExistingData(true);
            } catch (error) {
                console.error('Error fetching prediction data:', error);
            }
        };
        fetchPredictionAssessments();

        const fetchAverageType = async () => {
            try {
                const fetchAverageType = await axios.get(`${apiURL}/get_average_predicted_score_by_assessment_type`);
                console.log('Average Types:', fetchAverageType.data);
                setDataAverageType(fetchAverageType.data);
                setIsExistingDataAverageType(true);
            } catch (error) {
                console.error('Error fetching average types:', error);
            }
        };
        fetchAverageType();

        const fetchPredictionByAssessmentID = async () => {
            try {
                const fetch = await axios.get(`${apiURL}/get_student_count_by_assessment_id`);
                console.log('Assessments by ID:', fetch.data);
                setDataCountByAssessmentID(fetch.data);
                setIsExistingDataCountByAssessmentID(true);
            } catch (error) {
                console.error('Error fetching data by assessment ID:', error);
            }
        };
        fetchPredictionByAssessmentID();

        console.log('Data:', dataCourses, dataStudentRegistration, dataAssessments, dataVle, dataStudentInfo, dataStudentAssessment, dataStudentVle);

    }, []);



    return (
        <>
            <Grid container spacing={1}>
                <Grid item container spacing={1}>
                    <Box component="main" maxWidth="md" sx={{ mt: 10, p: 3, bgcolor: 'background.paper', boxShadow: 3, borderRadius: 2, minWidth: '100%', minHeight: '80%' }}>
                        <Grid container spacing={1}>
                            <Grid item container xs={3}>
                                {files.length > 0 ? (
                                    <Alert severity="info" sx={{ minHeight: 200, minWidth: '100%' }}>
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
                                    <Alert severity="error" sx={{ minHeight: 200, minWidth: '100%' }}>
                                        No files found
                                    </Alert>
                                )}
                            </Grid>
                            <Grid item container xs={3}>
                                <Alert severity="info" sx={{ minHeight: 200, minWidth: '100%' }}>
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
                            <Grid item container xs={6}>
                                {isExistingData ? (
                                    <Alert severity="info" sx={{ minWidth: '100%' }}>
                                        <AlertTitle>Student Assessment Prediction</AlertTitle>
                                        <Grid container spacing={2}>
                                            <ScorePieChartComponent data={dataStudentAssessmentPrediction} />
                                        </Grid>
                                    </Alert>
                                ) : (
                                    <Alert severity="error" sx={{ minHeight: 200, minWidth: '100%' }}>
                                        No prediction data found
                                    </Alert>
                                )}
                            </Grid>
                        </Grid>



                    </Box>
                </Grid>
                <Grid item container spacing={1}>
                    <Box component="main" maxWidth="md" sx={{ mt: 2, p: 3, bgcolor: 'background.paper', boxShadow: 3, borderRadius: 2, minWidth: '100%', minHeight: '80%' }}>
                        <Grid container spacing={1}>
                            <Grid item container xs={6}>
                                {isExistingDataAverageType ? (
                                    <Alert severity="info" sx={{ minHeight: 200, minWidth: '100%' }}>
                                        <AlertTitle>{dataAverageType.length} Types of Assessments found in <b>{new Intl.NumberFormat('en-US').format(dataStudentAssessment * dataAssessments)}</b> documents...</AlertTitle>
                                        <Grid container spacing={2}>
                                            <TableContainer component={Box}>
                                                <Table>
                                                    <TableHead>
                                                        <TableRow>
                                                            <TableCell>Assessment Type</TableCell>
                                                            <TableCell>Description</TableCell>
                                                            <TableCell>Average Score</TableCell>
                                                            <TableCell>Evaluation</TableCell>
                                                        </TableRow>
                                                    </TableHead>
                                                    <TableBody>
                                                        {dataAverageType.map((averageType, index) => (
                                                            <TableRow key={index}>
                                                                <TableCell>{averageType.AssessmentType}</TableCell>
                                                                <TableCell>
                                                                    {averageType.AssessmentType === 'TMA' ? 'Tutor Marked Assessment' :
                                                                        averageType.AssessmentType === 'Exam' ? 'Examination' :
                                                                            averageType.AssessmentType === 'CMA' ? 'Computer Marked Assessment' :
                                                                                averageType.AssessmentType === 'Coursework' ? 'Coursework' :
                                                                                    averageType.AssessmentType === 'Quiz' ? 'Quiz' : 'Unknown'}
                                                                </TableCell>
                                                                <TableCell><b>{averageType.AverageScore.toFixed(2)}</b></TableCell>
                                                                <TableCell sx={{ p: 0}}>
                                                                    {averageType.AverageScore < 70 ? (
                                                                        <Alert severity="error" sx={{ borderRadius: '25px'}}>Low</Alert>
                                                                    ) : averageType.AverageScore < 80 && averageType.AverageScore >= 70 ? (
                                                                        <Alert severity="warning" sx={{ borderRadius: '25px'}}>Medium</Alert>
                                                                    ) : averageType.AverageScore < 90 && averageType.AverageScore >= 80 ? (
                                                                        <Alert severity="info" sx={{ backgroundColor: 'primary.light', color: 'white', borderRadius: '25px' }}>High</Alert>
                                                                    ) : (
                                                                        <Alert severity="success" sx={{ borderRadius: '25px'}}>Excelent</Alert>
                                                                    )}
                                                                </TableCell>
                                                            </TableRow>
                                                        ))}
                                                    </TableBody>
                                                </Table>
                                            </TableContainer>
                                        </Grid>
                                    </Alert>
                                ) : (
                                    !isExistingData ? (
                                        <Alert severity="error" sx={{ minHeight: 200, minWidth: '100%' }}>
                                            No data average types found
                                        </Alert>
                                    ) : (
                                        <Alert severity="warning" sx={{ minHeight: 200, minWidth: '100%' }}>
                                            Processing data average from <b>{new Intl.NumberFormat('en-US').format(dataStudentAssessment * dataAssessments)}</b> documents...
                                        </Alert>
                                    )
                                )}
                            </Grid>
                            <Grid item container xs={6}>
                                {isExistingDataCountByAssessmentID ? (
                                    <Alert severity="info" sx={{ minWidth: '100%' }}>
                                        <AlertTitle>Student by Assessment ID</AlertTitle>
                                        <Grid container spacing={2} sx={{ minWidth: '100%' }}>
                                            <AssessmentByIdBarChartComponent data={dataCountByAssessmentID} />
                                        </Grid>
                                    </Alert>
                                ) : (
                                    <Alert severity="error" sx={{ minHeight: 200, minWidth: '100%' }}>
                                        No assessments by ID data found
                                    </Alert>
                                )}
                            </Grid>
                        </Grid>
                    </Box>
                </Grid>
            </Grid>
        </>

    );
};

export default MainContentComponent;
