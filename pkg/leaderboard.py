# -*- coding: utf-8 -*-
from __future__ import unicode_literals
import dotenv
import os
import math
import datetime
import calendar
import json
from redis import StrictRedis, Redis, ConnectionPool
from mongoengine.connection import get_db
from itertools import groupby
from django.conf import settings
from core.utils import awsConnectSession
from bson import ObjectId
from core.utils import ConnectionWrapper


def batch(iterable, n=0):
    l = len(iterable)
    for ndx in range(0, l, n):
        yield iterable[ndx:min(ndx+n, l)]


class LeaderBoard(object):
    DEFAULT_PAGE_SIZE = 100
    DEFAULT_REDIS_HOST = None
    DEFAULT_REDIS_PORT = None
    DEFAULT_REDIS_DB = 0
    DEFAULT_MEMBER_DATA_NAMESPACE = 'member_data'
    DEFAULT_GLOBAL_MEMBER_DATA = False
    DEFAULT_POOLS = {}
    ASC = 'asc'
    DESC = 'desc'
    MEMBER_KEY = 'n'
    MEMBER_DATA_KEY = 'member_data'
    SCORE_KEY = 's'
    RANK_KEY = 'r'
    rank_idx = 0
    leaderboard_name = None
    leaderboard_gameName = None
    prize_default = []
    rank_cutoff = 1000
    mongo_client = {}

    def get_mongo_collection(self, db_name):
        if not self.mongo_client:
            self.client = ConnectionWrapper().mongo_connect()
        db = self.client[settings.MONGO_DB]
        collection = db[db_name]
        return collection

    @classmethod
    def pool(self, host, port, db, pools={}, **options):
        '''
        Fetch a redis conenction pool for the unique combination of host
        and port. Will create a new one if there isn't one already.
        '''
        key = (host, port, db)
        rval = pools.get(key)
        if not isinstance(rval, ConnectionPool):
            rval = ConnectionPool(host=host, port=port, db=db, **options)
            pools[key] = rval
        return rval

    def __init__(self, **options):
        self.options = options

        self.page_size = self.options.pop('page_size', self.DEFAULT_PAGE_SIZE)
        if self.page_size < 1:
            self.page_size = self.DEFAULT_PAGE_SIZE

        self.member_data_namespace = self.options.pop(
            'member_data_namespace',
            self.DEFAULT_MEMBER_DATA_NAMESPACE)
        self.global_member_data = self.options.pop(
            'global_member_data',
            self.DEFAULT_GLOBAL_MEMBER_DATA)

        self.order = self.options.pop('order', self.DESC).lower()
        if not self.order in [self.ASC, self.DESC]:
            raise ValueError(
                "%s is not one of [%s]" % (self.order, ",".join([self.ASC, self.DESC])))

        redis_connection = self.options.pop('redis_connection', None)
        if redis_connection:
            # allow the developer to pass a raw redis connection and
            # we will use it directly instead of creating a new one
            self.redis_connection = redis_connection
        else:
            connection = self.options.pop('connection', None)
            if isinstance(connection, (StrictRedis, Redis)):
                self.options['connection_pool'] = connection.connection_pool
            if 'connection_pool' not in self.options:
                self.options['connection_pool'] = self.pool(
                    self.options.pop(
                        'host', settings.REDIS_URL_HOST or self.DEFAULT_REDIS_HOST),
                    self.options.pop(
                        'port', settings.REDIS_URL_PORT or self.DEFAULT_REDIS_PORT),
                    self.options.pop('db', self.DEFAULT_REDIS_DB),
                    self.options.pop('pools', self.DEFAULT_POOLS),
                    **self.options
                )
            self.redis_connection = Redis(**self.options)

    def batch(self, iterable, n=1):
        l = len(iterable)
        for ndx in range(0, l, n):
            raw_leader_data = self.redis_connection.zrevrange(
                self.leaderboard_name,
                int(ndx),
                int(ndx+n-1),
                withscores=True
            )
            yield (raw_leader_data, ndx)

    def calculate_rank(self, raw_leader_data, cursor):
        rank_list = []
        prize_list = {}

        idx = cursor  # index position
        # prize_default = (1,1000,2,300,3,100,5,75,10,50,100,30,500,20,1000,10)

        # if idx == 0:
        #     money_list.append([0, 0])
        for rank, (_, grp) in enumerate(groupby(raw_leader_data, key=lambda xs: xs[1]), 1):
            tie_rank = idx + 1
            for x in grp:
                if(tie_rank <= self.rank_cutoff):
                    # payouts
                    if tie_rank not in prize_list:
                        prize_list[tie_rank] = [0, 0]
                    prize_list[tie_rank][0] += 1

                    if self.rank_cutoff <= idx:
                        prize_list[tie_rank][1] += 0
                    else: 
                        prize_list[tie_rank][1] += float(
                            self.prize_default[self.rank_idx+1])
                        if(idx+1 - int(self.prize_default[self.rank_idx]) == 0):
                            self.rank_idx += 2
                rank_list.append(x + (tie_rank,))
                idx += 1
        return (rank_list, prize_list)


class CompetitionRanking(LeaderBoard):

    def get_offset(self, current_page, **options):
        page_size = options.get('page_size', self.page_size)

        index_for_redis = current_page - 1

        starting_offset = (index_for_redis * page_size)

        if starting_offset < 0:
            starting_offset = 0

        ending_offset = (starting_offset + page_size) + 1

        return (starting_offset, ending_offset)

    def get_created(self):
        dt = datetime.datetime.utcnow().replace(
            hour=16, minute=0, second=0, microsecond=0)
        return calendar.timegm(dt.timetuple())

    def bulk_insert(self, documents):
        bulk = get_db().leaderboards.initialize_unordered_bulk_op()
        for document in documents:
            bulk.insert({
                "name": document[0],
                "gameName": self.leaderboard_gameName,
                "score": document[1],
                "rank": document[2],
                "prize": document[3],
                "createdAt": self.get_created()
            })
        try:
            bulk.execute()
        except Exception as e:
            print e
        return True

    def calculate_prize_distribution(self, rank_list, money_list):
        for idx, member in enumerate(rank_list):
            if member[2] in money_list:
                count, amount = money_list[member[2]]
                calc = amount / count
                rank_list[idx] = rank_list[idx] + (calc, self.currency)
            else:
                rank_list[idx] = rank_list[idx] + (0,)
        return rank_list

    def store_leaderboard_to_s3(self, file_name, data):
        session = awsConnectSession()
        s3 = session.resource("s3")

        if settings.DEBUG:
            bucket_name = 'payout-dev-leaderboard'
        else:
            bucket_name = 'payout-leaderboard'

        obj = s3.Object(
            bucket_name,
            'raw/%s/%s' % (self.leaderboard_name, file_name)
        )

        return obj.put(Body=json.dumps(data))

    def get_leaderboards_to_process(self, league_id):
        return get_db().tcategories.find(
            {'state': 'LIVE',
             'name': {'$in': ['versus', 'score']}
             })

    def get_tournament(self):
        return self.get_mongo_collection('tcategories').find_one(
            {
                '_id': ObjectId(self.t_cat_id)
            }
        )

    def get_leader_board(self):
        return self.get_mongo_collection('leaderboards').find_one(
            {
                '_id': ObjectId(self.leaderboard_id),
                'tournamentId': ObjectId(self.t_cat_id),
                'state': 'LIVE'

            }
        )

    def update_league_state(self, league, state):
        return get_db().leaderboards.find_and_modify(
            query={'_id': league['_id']},
            update={"$set": {'state': state}},
            upsert=False, full_response=False
        )

    def initiate_payouts(self, league):
        pass

    # def confirmation_prompt(self):
    #     try:
    #         leagues = self.get_leaderboards_to_process()
    #         count = 0
    #         for league in leagues:
    #             count += 1

    #             print '%s  g%s:slbd:%s ' % (
    #                 league['name'], league['gameProfileId'], league['_id'],)

    #         response = raw_input(
    #             '\033[1;32mfound %d leagues would you like to process these leagues for payouts? [y/N]:\033[1;m' % count
    #         )

    #         if response.lower() != 'y':
    #             return False

    #         return True
    #     except KeyboardInterrupt:
    #         print '\n shutting down'

    def get_rank(self, t_cat_id, leaderboard_id):
        # confirm = self.confirmation_prompt()
        # if not confirm:
        #     return

        if not t_cat_id:
            return

        self.t_cat_id = t_cat_id
        self.leaderboard_id = leaderboard_id
        league = self.get_tournament()
        leader_board = self.get_leader_board()

        if leader_board:
            self.update_league_state(leader_board, 'PROCESS')
            self.leaderboard_name = 'lb:%s' % leader_board['_id']
            self.prize_default = league['prizePool']
            self.rank_cutoff = int(league['prizePool'][-2])
            self.currency = league['currency']

            total_leaders = self.redis_connection.zcount(
                self.leaderboard_name, '-inf', '+inf')
            #chunk_size = self.rank_cutoff
            chunk_size = 5000
            self.rank_idx = 0

            for raw_leader_data, cursor in self.batch(range(0, total_leaders), chunk_size):
                rank_list, prize_list = self.calculate_rank(
                    raw_leader_data, cursor)
                prized_rank_list = self.calculate_prize_distribution(
                    rank_list, prize_list)

                if cursor == 0:
                    self.store_leaderboard_to_s3(
                        'lb-%d.json' % cursor, prized_rank_list)
            self.update_league_state(league, 'DONE')

            return (league, self.rank_idx)
        else:
            return (False, 0)


def run(*args):
    pass
    # if 'league_id' in args:
    leaderboard = CompetitionRanking()
    leaderboard.get_rank('5bb1f32f6e723308200654d3',
                         '5bb37a866fde6f2b1f2aa3b6')
