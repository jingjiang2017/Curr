import request from '../utils/request';

export async function addUQuiz(params) {
  return request('/api/quizs',{
    method: 'post',
    body: JSON.stringify(params),
  });
}

export async function getAllLotterys(params) {
  return request(`/api/stat/lotterys?pageSize=${params.pageSize||10}&currPage=${params.currPage||1}&no=${params.no||''}`,{
    method: 'get',
  });
}

export async function getAllBrokerages(params) {
  return request(`/api/stat/brokerages?pageSize=${params.pageSize||10}&currPage=${params.currPage||1}&createdAt=${params.createdAt||''}`,{
    method: 'get',
  });
}

export async function getAllUserStats(params) {
  return request(`/api/stat/users?pageSize=${params.pageSize || 10}&currPage=${params.currPage || 1}&startTime=${params.startTime || ''}&endTime=${params.endTime || ''}`, {
    method: 'get',
  });
}


